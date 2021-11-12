package config

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/go-kit/kit/log/level"
)

type Config struct {
	Name string

	LogFormat string
	LogLevel  level.Option

	Server ServerConfig

	BucketConfig []byte
}

type ServerConfig struct {
	Listen         string
	ListenInternal string
	HealthcheckURL string
}

func ParseFlags() (*Config, error) {
	cfg := &Config{} //nolint:exhaustivestruct

	// Logger flags.
	flag.StringVar(&cfg.Name, "debug.name", "rules-objstore", "A name to add as a prefix to log lines.")
	logLevelRaw := flag.String("log.level", "info", "The log filtering level. Options: 'error', 'warn', 'info', 'debug'.")
	flag.StringVar(&cfg.LogFormat, "log.format", "logfmt", "The log format to use. Options: 'logfmt', 'json'.")

	// Server flags.
	flag.StringVar(&cfg.Server.Listen, "web.listen", ":8080", "The address on which the public server listens.")
	flag.StringVar(&cfg.Server.ListenInternal, "web.internal.listen", ":8081", "The address on which the internal server listens.")
	flag.StringVar(&cfg.Server.HealthcheckURL, "web.healthchecks.url", "http://localhost:8080", "The URL against which to run healthchecks.")

	// Object Storage flags.
	objstoreConfigPath := flag.String(
		"objstore.config-file",
		"",
		"YAML file that contains object store configuration. See format details: https://thanos.io/tip/thanos/storage.md/#configuration.",
	)

	flag.Parse()

	ll, err := parseLogLevel(logLevelRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	cfg.LogLevel = ll

	objstoreConfig, err := ioutil.ReadFile(*objstoreConfigPath)
	if err != nil {
		return nil, fmt.Errorf("invalid objstore config file path: %w", err)
	}

	cfg.BucketConfig = objstoreConfig

	return cfg, nil
}

func parseLogLevel(logLevelRaw *string) (level.Option, error) {
	switch *logLevelRaw {
	case "error":
		return level.AllowError(), nil
	case "warn":
		return level.AllowWarn(), nil
	case "info":
		return level.AllowInfo(), nil
	case "debug":
		return level.AllowDebug(), nil
	default:
		return nil, fmt.Errorf("unexpected log level: %s", *logLevelRaw) //nolint:goerr113
	}
}
