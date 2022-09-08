IMG := kfp-operator-argo-kfp-compiler

all: build

##@ Development

PENGUIN_STAGING := $(shell mktemp -d)

penguin_staging:
	cp ../../docs-gen/includes/quickstart/penguin_pipeline/*.py $(PENGUIN_STAGING)
	cp integration/Dockerfile $(PENGUIN_STAGING)

penguin_37: penguin_staging
	echo $(PENGUIN_STAGING)
	docker build --build-arg PYTHON_VERSION=3.7.12 --build-arg TFX_VERSION=1.9.1 -t penguin:3.7 $(PENGUIN_STAGING)

penguin_39: penguin_staging
	docker build --build-arg PYTHON_VERSION=3.9.13 --build-arg TFX_VERSION=1.9.1 -t penguin:3.9 $(PENGUIN_STAGING)

integration-test: docker-build penguin_37 penguin_39
	$(eval TMP := $(shell mktemp -d))
	docker run -v $(TMP):/shared kfp-operator-argo-kfp-compiler /shared
	docker run -v $(TMP):/shared --entrypoint /shared/compile.sh penguin:3.7 --output_file /tmp/pipeline.yaml --pipeline_config '{"rootLocation": "pipeline_root", "servingLocation": "serving", "name": "test", "image": "test-pipeline", "tfxComponents": "pipeline.create_components"}'
	docker run -v $(TMP):/shared --entrypoint /shared/compile.sh penguin:3.9 --output_file /tmp/pipeline.yaml --pipeline_config '{"rootLocation": "pipeline_root", "servingLocation": "serving", "name": "test", "image": "test-pipeline", "tfxComponents": "pipeline.create_components"}'

##@ Build

build:
	pip install poetry-dynamic-versioning --quiet
	poetry install
	poetry build

##@ Containers

WHEEL_VERSION=$(shell poetry version | cut -d ' ' -f 2)
DOCKER_BUILD_EXTRA_PARAMS=--build-arg WHEEL_VERSION=${WHEEL_VERSION}
include ../../docker-targets.mk
