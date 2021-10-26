package server

import (
	"io"
	"net/http"

	rulesspec "github.com/observatorium/api/rulesbackend/server/v1"
	"github.com/thanos-io/objstore"
)

const (
	rulesPath = "/rules.yaml"
)

type Server struct {
	bucket objstore.Bucket
}

func NewServer(bucket objstore.Bucket) *Server {
	return &Server{
		bucket: bucket,
	}
}

var _ rulesspec.ServerInterface = &Server{}

func (s *Server) ListRules(w http.ResponseWriter, r *http.Request, tenant string) {
	file, err := s.bucket.Get(r.Context(), tenant+rulesPath)
	defer file.Close()

	if err != nil {
		if s.bucket.IsObjNotFoundErr(err) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("no rules file found"))

			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something wrong happened"))

		return
	}

	w.Header().Set("Content-Type", "application/yaml")
	io.Copy(w, file)
}

func (s *Server) SetRules(w http.ResponseWriter, r *http.Request, tenant string) {
	err := s.bucket.Upload(r.Context(), tenant+rulesPath, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something wrong happened"))

		return
	}

	w.Write([]byte("successfully updated rules file"))
}
