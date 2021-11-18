FROM golang:1.17.3-alpine3.14 as builder

RUN apk add --update --no-cache ca-certificates tzdata git make bash && update-ca-certificates

ADD . /opt
WORKDIR /opt

RUN git update-index --refresh; make build

FROM alpine:3.14 as runner

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /opt/rules-objstore /bin/rules-objstore

ARG BUILD_DATE
ARG VERSION
ARG VCS_REF
ARG DOCKERFILE_PATH

LABEL vendor="Observatorium" \
    name="observatorium/rules-objstore" \
    description="Rules Object Storage Backend" \
    io.k8s.display-name="observatorium/rules-objstore" \
    io.k8s.description="Rules Object Storage Backend" \
    maintainer="Observatorium <team-monitoring@redhat.com>" \
    version="$VERSION" \
    org.label-schema.build-date=$BUILD_DATE \
    org.label-schema.description="Rules Object Storage Backend" \
    org.label-schema.docker.cmd="docker run --rm observatorium/rules-objstore" \
    org.label-schema.docker.dockerfile=$DOCKERFILE_PATH \
    org.label-schema.name="observatorium/rules-objstore" \
    org.label-schema.schema-version="1.0" \
    org.label-schema.vcs-branch=$VCS_BRANCH \
    org.label-schema.vcs-ref=$VCS_REF \
    org.label-schema.vcs-url="https://github.com/observatorium/rules-objstore" \
    org.label-schema.vendor="observatorium/rules-objstore" \
    org.label-schema.version=$VERSION

ENTRYPOINT ["/bin/rules-objstore"]
