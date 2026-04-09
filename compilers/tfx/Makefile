include ../../help.mk

##@ Build

lock: ## Update uv.lock after changing pyproject.toml
	uv lock

build: ## Build
	uv sync --locked --all-extras
	uv build

##@ Containers

include ../../docker-targets.mk

define penguin-staging
cp ../../docs-gen/includes/master/quickstart/penguin_pipeline/*.py $$STAGING_DIR && \
cp integration/Dockerfile integration/pipeline.yaml $$STAGING_DIR
endef

penguin-39:
	STAGING_DIR=$$(mktemp -d) && \
	$(penguin-staging) && \
	cp integration/penguin-39/pyproject.toml integration/penguin-39/uv.lock $$STAGING_DIR && \
	docker build --build-arg PYTHON_VERSION=3.9.13 --build-arg UV_VERSION=${UV_VERSION} -t penguin:3.9 $$STAGING_DIR

penguin-310:
	STAGING_DIR=$$(mktemp -d) && \
	$(penguin-staging) && \
	cp integration/penguin-310/pyproject.toml integration/penguin-310/uv.lock $$STAGING_DIR && \
	docker build --build-arg PYTHON_VERSION=3.10.12 --build-arg UV_VERSION=${UV_VERSION} -t penguin:3.10 $$STAGING_DIR

penguin-310-115:
	STAGING_DIR=$$(mktemp -d) && \
	$(penguin-staging) && \
	cp integration/penguin-310-115/pyproject.toml integration/penguin-310-115/uv.lock $$STAGING_DIR && \
	docker build --build-arg PYTHON_VERSION=3.10.12 --build-arg UV_VERSION=${UV_VERSION} -t penguin:3.10-115 $$STAGING_DIR

IMG := kfp-operator-tfx-compiler
WHEEL_VERSION=$(shell uv run python -c "import compiler; print(compiler.__version__)" 2>/dev/null || echo "0.0.0")
UV_VERSION=$(shell grep '^uv ' ../../.tool-versions | awk '{print $$2}')
DOCKER_BUILD_EXTRA_PARAMS=--build-arg WHEEL_VERSION=${WHEEL_VERSION} --build-arg UV_VERSION=${UV_VERSION} --build-context base-common="../resources"

##@ Development

test: build ## Run all tests
	uv run pytest

penguin-images: penguin-39 penguin-310 penguin-310-115

integration-test: docker-build ## Run the integration test
	$(MAKE) -j3 penguin-images
	docker volume create tfx-integration-test-shared; \
	trap 'docker volume rm tfx-integration-test-shared' EXIT; \
	docker run --rm -v tfx-integration-test-shared:/shared alpine chmod 777 /shared && \
	docker run --rm -v tfx-integration-test-shared:/shared kfp-operator-tfx-compiler /shared && \
	( \
	  docker run --rm -v tfx-integration-test-shared:/shared --entrypoint /shared/compile.sh penguin:3.9 --output_file /shared/pipeline_v3.9.json --pipeline_config /pipeline/pipeline.yaml & PID39=$$!; \
	  docker run --rm -v tfx-integration-test-shared:/shared --entrypoint /shared/compile.sh penguin:3.10 --output_file /shared/pipeline_v3.10.json --pipeline_config /pipeline/pipeline.yaml & PID310=$$!; \
	  docker run --rm -v tfx-integration-test-shared:/shared --entrypoint /shared/compile.sh penguin:3.10-115 --output_file /shared/pipeline_v3.10-115.json --pipeline_config /pipeline/pipeline.yaml & PID310_115=$$!; \
	  wait $$PID39 && wait $$PID310 && wait $$PID310_115 \
	)
