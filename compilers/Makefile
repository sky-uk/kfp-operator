include ../docker-targets.mk
include ../help.mk

IMG := kfp-operator-tfx-compiler

all: build

##@ Development

test: ## Test
	$(MAKE) -C tfx test
##@ Build

build: ## Build
	$(MAKE) -C tfx build

##@ Containers

WHEEL_VERSION=$(shell poetry version | cut -d ' ' -f 2)
DOCKER_BUILD_EXTRA_PARAMS=--build-arg WHEEL_VERSION=${WHEEL_VERSION}

docker-push-all: ## Push all compiler docker images
	$(MAKE) docker-push IMG=kfp-operator-tfx-compiler
