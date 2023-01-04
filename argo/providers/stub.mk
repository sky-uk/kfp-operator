include ../../common.mk

IMG := kfp-operator-stub-provider

##@ Development

test:
	go test ./... -tags=unit

##@ Build

build:
	go build -o bin/provider ./stub/cmd

##@ Containers

DOCKER_BUILD_EXTRA_PARAMS=-f stub/Dockerfile
include ../../docker-targets.mk
