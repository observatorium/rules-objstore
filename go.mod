module github.com/observatorium/rules-objstore

go 1.16

replace (
	github.com/efficientgo/tools/core v0.0.0-unpublish => ../../efficientgo/tools/core
	// TODO: Remove this: https://github.com/thanos-io/thanos/issues/3967.
	github.com/minio/minio-go/v7 => github.com/bwplotka/minio-go/v7 v7.0.11-0.20210324165441-f9927e5255a6
	github.com/observatorium/api/rulesbackend/server v0.0.0-unpublish => ../api/rulesbackend/server
	github.com/thanos-io/objstore v0.0.0-unpublish => ../../thanos/objstore
)

require (
	github.com/go-kit/kit v0.11.0
	github.com/metalmatze/signal v0.0.0-20210307161603-1c9aa721a97a
	github.com/observatorium/api/rulesbackend/server v0.0.0-unpublish
	github.com/oklog/run v1.0.0
	github.com/prometheus/client_golang v1.11.0
	github.com/thanos-io/objstore v0.0.0-unpublish
)
