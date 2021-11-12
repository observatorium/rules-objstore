package main

import (
	"context"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/metalmatze/signal/healthcheck"
	"github.com/metalmatze/signal/internalserver"
	rulesspec "github.com/observatorium/api/rules"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/thanos-io/thanos/pkg/objstore/client"

	"github.com/observatorium/rules-objstore/pkg/config"
	"github.com/observatorium/rules-objstore/pkg/server"
)

func main() {
	cfg, err := config.ParseFlags()
	if err != nil {
		stdlog.Fatal(err)
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	if cfg.LogFormat == "json" {
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	}

	logger = level.NewFilter(logger, cfg.LogLevel)
	if cfg.Name != "" {
		logger = log.With(logger, "name", cfg.Name)
	}

	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	defer level.Info(logger).Log("msg", "exiting")

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		//nolint:exhaustivestruct
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	healthchecks := healthcheck.NewMetricsHandler(healthcheck.NewHandler(), reg)

	if cfg.Server.HealthcheckURL != "" {
		// Checks if server is up.
		healthchecks.AddLivenessCheck("http",
			healthcheck.HTTPCheck(
				cfg.Server.HealthcheckURL,
				http.MethodGet,
				http.StatusNotFound,
				time.Second,
			),
		)
	}

	bkt, err := client.NewBucket(logger, cfg.BucketConfig, reg, cfg.Name)
	if err != nil {
		stdlog.Fatalf("creating object store bucket client: %v", err)
	}

	level.Info(logger).Log("msg", "starting rules-objstore")

	var g run.Group
	{
		// Signal channels must be buffered.
		sig := make(chan os.Signal, 1)
		g.Add(func() error {
			signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
			<-sig
			level.Info(logger).Log("msg", "caught interrupt")

			return nil
		}, func(_ error) {
			close(sig)
		})
	}
	{
		//nolint:exhaustivestruct
		s := http.Server{
			Addr: cfg.Server.Listen,
			Handler: rulesspec.Handler(
				server.NewServer(bkt, log.With(logger, "component", "server")),
			),
		}

		g.Add(func() error {
			level.Info(logger).Log("msg", "starting the HTTP server", "address", cfg.Server.Listen)

			return s.ListenAndServe() //nolint:wrapcheck
		}, func(_ error) {
			level.Info(logger).Log("msg", "shutting down the HTTP server")
			_ = s.Shutdown(context.Background())
		})
	}

	{
		h := internalserver.NewHandler(
			internalserver.WithName("Internal - rules-objstore API"),
			internalserver.WithHealthchecks(healthchecks),
			internalserver.WithPrometheusRegistry(reg),
			internalserver.WithPProf(),
		)

		//nolint:exhaustivestruct
		s := http.Server{
			Addr:    cfg.Server.ListenInternal,
			Handler: h,
		}

		g.Add(func() error {
			level.Info(logger).Log("msg", "starting internal HTTP server", "address", s.Addr)

			return s.ListenAndServe() //nolint:wrapcheck
		}, func(_ error) {
			_ = s.Shutdown(context.Background())
		})
	}

	if err := g.Run(); err != nil {
		stdlog.Fatal(err)
	}
}
