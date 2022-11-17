IMG := kfp-operator-vai-provider

##@ Development

test:
	@echo no tests defined

##@ Build

build:
	go build -o bin/provider ./vai

##@ Containers

DOCKER_BUILD_EXTRA_PARAMS=-f vai/Dockerfile
include ../../docker-targets.mk
