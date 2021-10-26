module github.com/observatorium/rules-objstore

go 1.17

replace (
	github.com/efficientgo/tools/core v0.0.0-unpublish => ../../efficientgo/tools/core
	// TODO: Remove this: https://github.com/thanos-io/thanos/issues/3967.
	github.com/minio/minio-go/v7 => github.com/bwplotka/minio-go/v7 v7.0.11-0.20210324165441-f9927e5255a6
	github.com/observatorium/api/rulesbackend/server v0.0.0-unpublish => ../api/rulesbackend/server
	github.com/thanos-io/objstore v0.0.0-unpublish => ../../thanos/objstore
)

require (
	github.com/observatorium/api/rulesbackend/server v0.0.0-unpublish
	github.com/thanos-io/objstore v0.0.0-unpublish
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deepmap/oapi-codegen v1.8.3 // indirect
	github.com/efficientgo/tools/core v0.0.0-unpublish // indirect
	github.com/go-chi/chi/v5 v5.0.0 // indirect
	github.com/go-kit/kit v0.11.0 // indirect
	github.com/go-logfmt/logfmt v0.5.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.30.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	go.uber.org/goleak v1.1.10 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/sys v0.0.0-20210823070655-63515b42dcdf // indirect
	golang.org/x/tools v0.1.5 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)
