include ../../help.mk

##@ Build

lock: ## Update uv.lock after changing pyproject.toml
	uv lock

build: ## Build
	uv sync --locked --all-extras
	uv run ruff format --check .
	uv run ruff check .
	uv build

##@ Containers

include ../../docker-targets.mk
IMG := kfp-operator-kfpsdk-compiler
WHEEL_VERSION=$(shell uv run python -c "import compiler; print(compiler.__version__)" 2>/dev/null || echo "0.0.0")
UV_VERSION=$(shell grep '^uv ' ../../.tool-versions | awk '{print $$2}')
DOCKER_BUILD_EXTRA_PARAMS=--build-arg WHEEL_VERSION=${WHEEL_VERSION} --build-arg UV_VERSION=${UV_VERSION} --build-context base-common="../resources"

define quickstart-staging
cp -r ../../docs-gen/includes/master/kfpsdk-quickstart/getting_started $$STAGING_DIR/ && \
cp integration/Dockerfile integration/pipeline.yaml $$STAGING_DIR
endef

quickstart-39:
	STAGING_DIR=$$(mktemp -d) && \
	$(quickstart-staging) && \
	cp integration/quickstart-39/pyproject.toml integration/quickstart-39/uv.lock $$STAGING_DIR && \
	docker build --build-arg PYTHON_VERSION=3.9.21-slim --build-arg UV_VERSION=${UV_VERSION} -t kfpsdk-quickstart:3.9 $$STAGING_DIR

quickstart-310:
	STAGING_DIR=$$(mktemp -d) && \
	$(quickstart-staging) && \
	cp integration/quickstart-310/pyproject.toml integration/quickstart-310/uv.lock $$STAGING_DIR && \
	docker build --build-arg PYTHON_VERSION=3.10-slim --build-arg UV_VERSION=${UV_VERSION} -t kfpsdk-quickstart:3.10 $$STAGING_DIR

quickstart-311:
	STAGING_DIR=$$(mktemp -d) && \
	$(quickstart-staging) && \
	cp integration/quickstart-311/pyproject.toml integration/quickstart-311/uv.lock $$STAGING_DIR && \
	docker build --build-arg PYTHON_VERSION=3.11-slim --build-arg UV_VERSION=${UV_VERSION} -t kfpsdk-quickstart:3.11 $$STAGING_DIR

quickstart-docs:
	$(MAKE) -C ../../docs-gen/includes/master/kfpsdk-quickstart docker-build

##@ Development

format: ## Format and fix source files in place
	uv run ruff format .
	uv run ruff check --fix .

test: build ## Run unit tests
	uv run pytest

quickstart-images: quickstart-39 quickstart-310 quickstart-311 quickstart-docs

integration-test: docker-build ## Build and run integration tests
	$(MAKE) -j4 quickstart-images
	docker volume create kfpsdk-integration-test-shared; \
	trap 'docker volume rm kfpsdk-integration-test-shared' EXIT; \
	docker run --rm -v kfpsdk-integration-test-shared:/shared alpine chmod 777 /shared && \
	docker run --rm -v kfpsdk-integration-test-shared:/shared ${IMG} /shared && \
	( \
	  docker run --rm -v kfpsdk-integration-test-shared:/shared --entrypoint /shared/compile.sh kfpsdk-quickstart:3.9 --output_file /shared/pipeline_v3.9.json --pipeline_config /pipeline/pipeline.yaml & PID39=$$!; \
	  docker run --rm -v kfpsdk-integration-test-shared:/shared --entrypoint /shared/compile.sh kfpsdk-quickstart:3.10 --output_file /shared/pipeline_v3.10.json --pipeline_config /pipeline/pipeline.yaml & PID310=$$!; \
	  docker run --rm -v kfpsdk-integration-test-shared:/shared --entrypoint /shared/compile.sh kfpsdk-quickstart:3.11 --output_file /shared/pipeline_v3.11.json --pipeline_config /pipeline/pipeline.yaml & PID311=$$!; \
	  docker run --rm -v kfpsdk-integration-test-shared:/shared -v ./integration:/pipeline --entrypoint /shared/compile.sh kfp-operator-kfpsdk-quickstart:${VERSION} --output_file /shared/pipeline_v3.12.json --pipeline_config /pipeline/pipeline.yaml & PID312=$$!; \
	  wait $$PID39 && wait $$PID310 && wait $$PID311 && wait $$PID312 \
	)
