include ../../../common.mk
include ../../../docker-targets.mk

IMG := kfp-operator-kfp-provider

all: build

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/provider ./cmd

unit-test:
	go test ./... -tags=unit

test: unit-test
