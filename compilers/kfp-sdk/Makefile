include ../../help.mk

##@ Build

build: ## Build
	uv sync --all-extras
	uv run black .
	uv run flake8
	uv build

##@ Containers

include ../../docker-targets.mk
IMG := kfp-operator-kfpsdk-compiler
WHEEL_VERSION=$(shell uv run python -c "import compiler; print(compiler.__version__)" 2>/dev/null || echo "0.0.0")
DOCKER_BUILD_EXTRA_PARAMS=--build-arg WHEEL_VERSION=${WHEEL_VERSION} --build-context base-common=".."

##@ Development

test: build ## Run unit tests
	uv run pytest

integration-test: docker-build ## Build and run integration tests
	$(MAKE) -C ../../docs-gen/includes/master/kfpsdk-quickstart docker-build
	$(eval TMP := $(shell mktemp -d))
	docker run --user $(id -u):$(id -g) -v $(TMP):/shared ${IMG}:${VERSION} /shared
	docker run --user $(id -u):$(id -g) -v $(TMP):/shared -v ./integration:/pipeline --entrypoint /shared/compile.sh kfp-operator-kfpsdk-quickstart:${VERSION} --pipeline_config /pipeline/pipeline.yaml --output_file /shared/out.json
	rm -rf $(TMP)
