module github.com/observatorium/rules-objstore

go 1.16

replace (
	// TODO: Remove this: https://github.com/thanos-io/thanos/issues/3967.
	github.com/minio/minio-go/v7 => github.com/bwplotka/minio-go/v7 v7.0.11-0.20210324165441-f9927e5255a6

	github.com/thanos-io/thanos v0.22.0 => github.com/thanos-io/thanos v0.19.1-0.20211112050938-aa7e9f33aaf0
)

require (
	github.com/baiyubin/aliyun-sts-go-sdk v0.0.0-20180326062324-cfa1a18b161f // indirect
	github.com/go-kit/kit v0.11.0
	github.com/metalmatze/signal v0.0.0-20210307161603-1c9aa721a97a
	github.com/observatorium/api v0.1.3-0.20211112102146-7e7baedacb84
	github.com/oklog/run v1.1.0
	github.com/prometheus/client_golang v1.11.0
	github.com/thanos-io/thanos v0.22.0
)
