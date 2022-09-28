include .bingo/Variables.mk

SHELL=/usr/bin/env bash -o pipefail
FIRST_GOPATH := $(firstword $(subst :, ,$(shell go env GOPATH)))
OS ?= $(shell uname -s | tr '[A-Z]' '[a-z]')
ARCH ?= $(shell uname -m)

VERSION := $(strip $(shell [ -d .git ] && git describe --always --tags --dirty))
BUILD_DATE := $(shell date -u +"%Y-%m-%d")
BUILD_TIMESTAMP := $(shell date -u +"%Y-%m-%dT%H:%M:%S%Z")
VCS_BRANCH := $(strip $(shell git rev-parse --abbrev-ref HEAD))
VCS_REF := $(strip $(shell [ -d .git ] && git rev-parse --short HEAD))
DOCKER_REPO ?= quay.io/observatorium/rules-objstore

BIN_NAME ?= rules-objstore

default: $(BIN_NAME)
all: clean lint test $(BIN_NAME)

.PHONY: deps
deps: go.mod go.sum
	go mod tidy -compat=1.17
	go mod download
	go mod verify

$(BIN_NAME): deps main.go $(wildcard *.go) $(wildcard */*.go)
	CGO_ENABLED=0 GO111MODULE=on GOPROXY=https://proxy.golang.org go build -a -ldflags '-s -w' -o $@ .

.PHONY: build
build: $(BIN_NAME)

.PHONY: format
format: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --fix --enable-all -c .golangci.yml

.PHONY: go-fmt
go-fmt: $(GOFUMPT)
	@fmt_res=$$($(GOFUMPT) -l -w $$(find . -type f -name '*.go')); if [ -n "$$fmt_res" ]; then printf '\nGofmt found style issues. Please check the reported issues\nand fix them if necessary before submitting the code for review:\n\n%s' "$$fmt_res"; exit 1; fi

.PHONY: lint
lint: $(GOLANGCI_LINT) go-fmt
	$(GOLANGCI_LINT) run -v --enable-all -c .golangci.yml

.PHONY: test
test: build test-unit

.PHONY: test-unit
test-unit:
	CGO_ENABLED=1 GO111MODULE=on go test -v -race -short ./...

.PHONY: test-e2e
test-e2e: container-test
	CGO_ENABLED=1 GO111MODULE=on go test -v -race -short --tags integration ./test/e2e

.PHONY: clean
clean:
	-rm $(BIN_NAME)

.PHONY: manifests
manifests: jsonnet/example/manifests

jsonnet/example/manifests: jsonnet/example/main.jsonnet $(JSONNET) $(GOJSONTOYAML)
	-rm -rf jsonnet/example/manifests
	-mkdir jsonnet/example/manifests
	$(JSONNET) -m jsonnet/example/manifests jsonnet/example/main.jsonnet | xargs -I{} sh -c 'cat {} | $(GOJSONTOYAML) > {}.yaml' -- {}
	find jsonnet/example/manifests -type f ! -name '*.yaml' -delete

.PHONY: container
container: Dockerfile
	@docker build --build-arg BUILD_DATE="$(BUILD_TIMESTAMP)" \
		--build-arg VERSION="$(VERSION)" \
		--build-arg VCS_REF="$(VCS_REF)" \
		--build-arg VCS_BRANCH="$(VCS_BRANCH)" \
		--build-arg DOCKERFILE_PATH="/Dockerfile" \
		-t $(DOCKER_REPO):$(VCS_BRANCH)-$(BUILD_DATE)-$(VERSION) .
	@docker tag $(DOCKER_REPO):$(VCS_BRANCH)-$(BUILD_DATE)-$(VERSION) $(DOCKER_REPO):latest

.PHONY: container-push
container-push: container
	docker push $(DOCKER_REPO):$(VCS_BRANCH)-$(BUILD_DATE)-$(VERSION)
	docker push $(DOCKER_REPO):latest

.PHONY: container-release
container-release: VERSION_TAG = $(strip $(shell [ -d .git ] && git tag --points-at HEAD))
container-release: container
	# https://git-scm.com/docs/git-tag#Documentation/git-tag.txt---points-atltobjectgt
	@docker tag $(DOCKER_REPO):$(VCS_BRANCH)-$(BUILD_DATE)-$(VERSION) $(DOCKER_REPO):$(VERSION_TAG)
	docker push $(DOCKER_REPO):$(VERSION_TAG)
	docker push $(DOCKER_REPO):latest

.PHONY: container-test
container-test: # Use 'shortcut' to build test image if on Linux, otherwise full build.
ifeq ($(OS), linux)
container-test: build
	@docker build \
		-f Dockerfile.e2e-test \
		-t $(DOCKER_REPO):local_e2e_test .
else
container-test:
	@docker build \
		-f Dockerfile \
		-t $(DOCKER_REPO):local_e2e_test .
endif
