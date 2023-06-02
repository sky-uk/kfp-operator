include ../../../common.mk

IMG := kfp-operator-stub-provider

##@ Build

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/provider ./cmd

##@ Containers

include ../../../docker-targets.mk
