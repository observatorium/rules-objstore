package server

import (
	"io"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	rulesspec "github.com/observatorium/api/rulesbackend/server/v1"
	"github.com/thanos-io/objstore"
)

const (
	rulesPath = "/rules.yaml"
)

type Server struct {
	bucket objstore.Bucket
	logger log.Logger
}

func NewServer(bucket objstore.Bucket, logger log.Logger) *Server {
	return &Server{
		bucket: bucket,
		logger: logger,
	}
}

var _ rulesspec.ServerInterface = &Server{}

func (s *Server) ListRules(w http.ResponseWriter, r *http.Request, tenant string) {
	logger := log.With(s.logger, "handler", "listrules", "tenant", tenant)

	file, err := s.bucket.Get(r.Context(), tenant+rulesPath)
	if err != nil {
		if s.bucket.IsObjNotFoundErr(err) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("no rules file found"))

			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("something wrong happened"))

		level.Warn(logger).Log("msg", "reading rules file from bucket", "err", err)

		return
	}

	defer file.Close()

	w.Header().Set("Content-Type", "application/yaml")

	if _, err := io.Copy(w, file); err != nil {
		level.Warn(logger).Log("msg", "copying rules file to HTTP response", "err", err)
	}
}

func (s *Server) SetRules(w http.ResponseWriter, r *http.Request, tenant string) {
	logger := log.With(s.logger, "handler", "setrules", "tenant", tenant)

	err := s.bucket.Upload(r.Context(), tenant+rulesPath, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("something wrong happened"))

		level.Warn(logger).Log("msg", "uploading rules file to bucket", "err", err)

		return
	}

	_, _ = w.Write([]byte("successfully updated rules file"))
}
