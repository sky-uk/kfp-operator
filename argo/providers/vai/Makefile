include ../../../common.mk
include ../../../docker-targets.mk

IMG := kfp-operator-vai-provider

##@ Development
unit-test:
	go test ./... -tags=unit

test: unit-test

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/provider ./cmd

