include ../../help.mk

##@ Build

build: ## Build
	pip install poetry-dynamic-versioning --quiet
	poetry install
	poetry run black .
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
	$(MAKE) -C ../../docs-gen/includes/master/kfpsdk-quickstart docker-build
	$(eval TMP := $(shell mktemp -d))
	docker run -v $(TMP):/shared ${IMG}:${VERSION} /shared
	docker run -v $(TMP):/shared -v ./integration:/pipeline --entrypoint /shared/compile.sh kfp-operator-kfpsdk-quickstart:${VERSION} --pipeline_config /pipeline/pipeline.yaml  --provider_config /pipeline/provider.json --output_file /shared/out.json
