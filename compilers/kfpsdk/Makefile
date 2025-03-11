##@ Build

build: ## Build
	pip install poetry-dynamic-versioning --quiet
	poetry install
	poetry build

##@ Containers

include ../../docker-targets.mk
IMG := kfp-operator-kfpsdk-compiler
WHEEL_VERSION=$(shell poetry version | cut -d ' ' -f 2)
DOCKER_BUILD_EXTRA_PARAMS=--build-arg WHEEL_VERSION=${WHEEL_VERSION} --build-context base-common=".."

##@ Development

test: ## Run unit tests
	poetry run pytest

integration-test: docker-build ## Build and run integration tests
	$(eval TMP := $(shell mktemp -d))
	docker run -v $(TMP):/shared kfp-operator-kfpsdk-compiler /shared
	docker run -v $(TMP):/shared --entrypoint /shared/compile.sh python:3.9
