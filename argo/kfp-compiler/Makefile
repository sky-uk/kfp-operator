include ../../help.mk

IMG := kfp-operator-argo-kfp-compiler

all: build

##@ Development

test: build ## Build and run all tests
	poetry run pytest

PENGUIN_STAGING := $(shell mktemp -d)

penguin-staging: ## Prepare the penguin image
	cp ../../docs-gen/includes/master/quickstart/penguin_pipeline/*.py $(PENGUIN_STAGING)
	cp integration/Dockerfile $(PENGUIN_STAGING)

penguin-38: penguin-staging ## Build the penguin image for Python 3.8
	docker build --build-arg PYTHON_VERSION=3.8.17 --build-arg TFX_VERSION=1.9.1 -t penguin:3.8 $(PENGUIN_STAGING)

penguin-39: penguin-staging ## Build the penguin image for Python 3.9
	docker build --build-arg PYTHON_VERSION=3.9.13 --build-arg TFX_VERSION=1.9.1 -t penguin:3.9 $(PENGUIN_STAGING)

penguin-310: penguin-staging ## Build the penguin image for Python 3.10
	docker build --build-arg PYTHON_VERSION=3.10.12 --build-arg TFX_VERSION=1.14.0 -t penguin:3.10 $(PENGUIN_STAGING)

integration-test: docker-build penguin-38 penguin-39 penguin-310 ## Run the integration test
	$(eval TMP := $(shell mktemp -d))
	echo '{"rootLocation": "pipeline_root", "servingLocation": "serving", "name": "test", "image": "test-pipeline", "tfxComponents": "pipeline.create_components"}' > $(TMP)/pipeline_config.json
	docker run -v $(TMP):/shared kfp-operator-argo-kfp-compiler /shared
	docker run -v $(TMP):/shared --entrypoint /shared/compile.sh penguin:3.10 --output_file /tmp/pipeline.yaml --pipeline_config /shared/pipeline_config.json  --execution_mode v1
	docker run -v $(TMP):/shared --entrypoint /shared/compile.sh penguin:3.8 --output_file /tmp/pipeline.yaml --pipeline_config /shared/pipeline_config.json --execution_mode v2
	docker run -v $(TMP):/shared --entrypoint /shared/compile.sh penguin:3.9 --output_file /tmp/pipeline.yaml --pipeline_config /shared/pipeline_config.json --execution_mode v2
	docker run -v $(TMP):/shared --entrypoint /shared/compile.sh penguin:3.10 --output_file /tmp/pipeline.yaml --pipeline_config /shared/pipeline_config.json  --execution_mode v2

##@ Build

build: ## Build
	pip install poetry-dynamic-versioning --quiet
	poetry install
	poetry build

##@ Containers

WHEEL_VERSION=$(shell poetry version | cut -d ' ' -f 2)
DOCKER_BUILD_EXTRA_PARAMS=--build-arg WHEEL_VERSION=${WHEEL_VERSION}

include ../../docker-targets.mk

docker-push-all: ## Push all compiler docker images
	$(MAKE) docker-push IMG=kfp-operator-argo-kfp-compiler
	$(MAKE) docker-push IMG=kfp-operator-tfx-compiler
