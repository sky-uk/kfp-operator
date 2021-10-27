include ../../version.mk

IMG := kfp-operator-argo-mlmc-cli

all: build

##@ Development

test: build
	poetry run pytest

##@ Build

build:
	pip install poetry-dynamic-versioning --quiet
	poetry install
	poetry build

##@ Containers

WHEEL_VERSION=$(shell poetry version | cut -d ' ' -f 2)
DOCKER_BUILD_EXTRA_PARAMS=--build-arg WHEEL_VERSION=${WHEEL_VERSION}
include ../../docker-targets.mk
