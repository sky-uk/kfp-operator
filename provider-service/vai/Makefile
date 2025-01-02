include ../../common.mk
include ../../docker-targets.mk

IMG := kfp-operator-vai-provider-service

##@ Development

unit-test:
	go test ./... -tags=unit

test: unit-test

##@ Build

build: generate
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/vai-provider-service cmd/main.go

##@ Containers

DOCKER_BUILD_EXTRA_PARAMS=--build-context base-source=".."
