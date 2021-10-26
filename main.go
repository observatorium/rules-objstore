package main

import (
	"net/http"

	rulesspec "github.com/observatorium/api/rulesbackend/server/v1"
	"github.com/thanos-io/objstore"

	"github.com/observatorium/rules-objstore/pkg/server"
)

func main() {
	http.ListenAndServe(
		":8080",
		rulesspec.Handler(
			server.NewServer(objstore.NewInMemBucket()),
		),
	)
}
