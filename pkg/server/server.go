package server

import (
	"io"
	"net/http"
	"path"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	rulesspec "github.com/observatorium/api/rules"
	"github.com/thanos-io/thanos/pkg/objstore"
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

// Make sure that Server implements rulesspec.ServerInterface.
var _ rulesspec.ServerInterface = &Server{} //nolint:exhaustivestruct

func (s *Server) ListRules(w http.ResponseWriter, r *http.Request, tenant string) {
	logger := log.With(s.logger, "handler", "listrules", "tenant", tenant)

	file, err := s.bucket.Get(r.Context(), getRulesFilePath(tenant))
	if err != nil {
		if s.bucket.IsObjNotFoundErr(err) {
			http.Error(w, "rules file not found", http.StatusNotFound)
			level.Debug(logger).Log("msg", "rules file not found", "path", getRulesFilePath(tenant))

			return
		}

		http.Error(w, "reading rules file from bucket", http.StatusInternalServerError)
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

	err := s.bucket.Upload(r.Context(), getRulesFilePath(tenant), r.Body)
	if err != nil {
		http.Error(w, "uploading rules file to bucket", http.StatusInternalServerError)
		level.Warn(logger).Log("msg", "uploading rules file to bucket", "err", err)

		return
	}

	_, _ = w.Write([]byte("successfully updated rules file"))
}

func getRulesFilePath(tenant string) string {
	return path.Join(tenant, rulesPath)
}
