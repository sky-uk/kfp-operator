include ../../../../../help.mk

##@ Build

build: ## Build
	poetry install
	poetry build

##@ Containers

include ../../../../../docker-targets.mk

IMG ?= kfp-operator-quickstart-dependant
WHEEL_VERSION=$(shell poetry version | cut -d ' ' -f 2)
DOCKER_BUILD_EXTRA_PARAMS=--build-arg WHEEL_VERSION=${WHEEL_VERSION}
