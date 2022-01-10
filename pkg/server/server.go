package server

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	rulesspec "github.com/observatorium/api/rules"
	"github.com/prometheus/prometheus/pkg/rulefmt"
	"github.com/thanos-io/thanos/pkg/objstore"
	"gopkg.in/yaml.v3"
)

const (
	rulesBasePath = "metrics/rules/"
	rulesFileName = "/rules.yaml"
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

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "reading request body", http.StatusInternalServerError)
		level.Warn(logger).Log("msg", "reading request body", "err", err)

		return
	}
	defer r.Body.Close()

	if _, errs := rulefmt.Parse(data); errs != nil {
		http.Error(w, "request body failed rule group validation", http.StatusBadRequest)
		level.Debug(logger).Log("msg", "request body failed rule group validation", "errs", errs)

		return
	}

	err = s.bucket.Upload(r.Context(), getRulesFilePath(tenant), bytes.NewReader(data))
	if err != nil {
		http.Error(w, "uploading rules file to bucket", http.StatusInternalServerError)
		level.Warn(logger).Log("msg", "uploading rules file to bucket", "err", err)

		return
	}

	_, _ = w.Write([]byte("successfully updated rules file"))
}

func (s *Server) ListAllRules(w http.ResponseWriter, r *http.Request) {
	logger := log.With(s.logger, "handler", "listAllRules")

	//nolint:exhaustivestruct
	allGroups := &rulefmt.RuleGroups{}

	if err := s.bucket.Iter(r.Context(), rulesBasePath, func(dir string) error {
		tenant := strings.TrimPrefix(dir, rulesBasePath)
		tenant = strings.TrimSuffix(tenant, "/")

		file, err := s.bucket.Get(r.Context(), getRulesFilePath(tenant))
		if err != nil {
			level.Warn(logger).Log("msg", "failed retrieving rules file", "tenant", tenant, "err", err)

			return err
		}
		defer file.Close()

		data, err := ioutil.ReadAll(file)
		if err != nil {
			level.Warn(logger).Log("msg", "error reading rules file", "tenant", tenant, "err", err)

			return err
		}

		groups, errs := rulefmt.Parse(data)
		if errs != nil {
			level.Warn(logger).Log("msg", "error parsing rules data", "tenant", tenant, "errs", errs)

			return err
		}

		// Append tenant name as prefix to the Rule group name to avoid duplicate group names across tenants.
		for _, rg := range groups.Groups {
			rg.Name = tenant + "." + rg.Name
			allGroups.Groups = append(allGroups.Groups, rg)
		}

		return nil
	}); err != nil {
		http.Error(w, "failed retrieving all rules", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/yaml")

	var buf bytes.Buffer
	if err := yaml.NewEncoder(&buf).Encode(allGroups); err != nil {
		http.Error(w, "marshalling rules to yaml", http.StatusInternalServerError)
		level.Warn(logger).Log("msg", "marshalling rules to yaml", "err", err)

		return
	}

	if _, err := w.Write(buf.Bytes()); err != nil {
		level.Warn(logger).Log("msg", "writing rules file to HTTP response", "err", err)
	}
}

func getRulesFilePath(tenant string) string {
	return path.Join(rulesBasePath, tenant, rulesFileName)
}
