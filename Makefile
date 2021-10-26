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

default: rules-objstore
all: clean lint test rules-objstore

.PHONY: deps
deps: go.mod go.sum
	go mod tidy
	go mod download
	go mod verify

rules-objstore: deps main.go $(wildcard *.go) $(wildcard */*.go)
	CGO_ENABLED=0 GO111MODULE=on GOPROXY=https://proxy.golang.org go build -a -ldflags '-s -w' -o $@ .

.PHONY: build
build: rules-objstore

.PHONY: format
format: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --fix --enable-all -c .golangci.yml

.PHONY: go-fmt
go-fmt: $(GOFUMPT)
	@fmt_res=$$(gofumpt -l -w $$(find . -type f -name '*.go')); if [ -n "$$fmt_res" ]; then printf '\nGofmt found style issues. Please check the reported issues\nand fix them if necessary before submitting the code for review:\n\n%s' "$$fmt_res"; exit 1; fi

.PHONY: lint
lint: $(GOLANGCI_LINT) go-fmt
	$(GOLANGCI_LINT) run -v --enable-all -c .golangci.yml

.PHONY: test
test: build test-unit

.PHONY: test-unit
test-unit:
	CGO_ENABLED=1 GO111MODULE=on go test -v -race -short ./...

.PHONY: clean
clean:
	-rm rules-objstore

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
