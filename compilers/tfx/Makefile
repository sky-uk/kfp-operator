include ../../help.mk

##@ Build

build: ## Build
	uv sync --all-extras
	uv build

##@ Containers

include ../../docker-targets.mk

PENGUIN_STAGING := $(shell mktemp -d)

penguin-staging: ## Prepare the penguin image
	cp ../../docs-gen/includes/master/quickstart/penguin_pipeline/*.py $(PENGUIN_STAGING)
	cp integration/* $(PENGUIN_STAGING)

penguin-39: penguin-staging ## Build the penguin image for Python 3.9
	docker build --build-arg PYTHON_VERSION=3.9.13 --build-arg TFX_VERSION=1.9.1 -t penguin:3.9 $(PENGUIN_STAGING)

penguin-310: penguin-staging ## Build the penguin image for Python 3.10
	docker build --build-arg PYTHON_VERSION=3.10.12 --build-arg TFX_VERSION=1.14.0 -t penguin:3.10 $(PENGUIN_STAGING)

penguin-310-115: penguin-staging ## Build the penguin image for Python 3.10
	docker build --build-arg PYTHON_VERSION=3.10.12 --build-arg TFX_VERSION=1.15.1 -t penguin:3.10-115 $(PENGUIN_STAGING)

IMG := kfp-operator-tfx-compiler
WHEEL_VERSION=$(shell uv run python -c "import compiler; print(compiler.__version__)" 2>/dev/null || echo "0.0.0")
DOCKER_BUILD_EXTRA_PARAMS=--build-arg WHEEL_VERSION=${WHEEL_VERSION} --build-context base-common=".."

##@ Development

test: build ## Run all tests
	uv run pytest

integration-test: docker-build penguin-39 penguin-310 penguin-310-115 ## Run the integration test
	$(eval TMP := $(shell mktemp -d))

	docker run -v $(TMP):/shared kfp-operator-tfx-compiler /shared

	docker run -v $(TMP):/shared --entrypoint /shared/compile.sh penguin:3.9 --output_file /shared/pipeline_v3.9.json --pipeline_config /pipeline/pipeline.yaml
	docker run -v $(TMP):/shared --entrypoint /shared/compile.sh penguin:3.10 --output_file /shared/pipeline_v3.10.json --pipeline_config /pipeline/pipeline.yaml
	docker run -v $(TMP):/shared --entrypoint /shared/compile.sh penguin:3.10-115 --output_file /shared/pipeline_v3.10-115.json --pipeline_config /pipeline/pipeline.yaml
