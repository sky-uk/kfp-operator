include ../../../../help.mk

##@ Build

build: ## Build
	uv sync
	uv build

##@ Containers

include ../../../../docker-targets.mk

IMG ?= kfp-operator-kfpsdk-quickstart
WHEEL_VERSION=$(shell uv run python -c "import importlib.metadata; print(importlib.metadata.version('getting_started'))" 2>/dev/null || echo "0.1.0")
DOCKER_BUILD_EXTRA_PARAMS=--build-arg WHEEL_VERSION=${WHEEL_VERSION}
