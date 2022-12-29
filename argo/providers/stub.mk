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
DOCKER_BUILD_EXTRA_TAGS=kfp-operator-argo-kfp-compiler kfp-operator-argo-kfp-pipeline
include ../../docker-targets.mk
