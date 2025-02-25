include ../../docker-targets.mk

IMG := kfp-operator-tfx-compiler
WHEEL_VERSION=0.0.0
DOCKER_BUILD_EXTRA_PARAMS=--build-arg WHEEL_VERSION=${WHEEL_VERSION}

build: ## Build
	uv build --package tfx_compiler

test: build ## Build and run all tests
	uv run pytest

PENGUIN_STAGING := $(shell mktemp -d)

penguin-staging: ## Prepare the penguin image
	cp ../../docs-gen/includes/master/quickstart/penguin_pipeline/*.py $(PENGUIN_STAGING)
	cp integration/Dockerfile $(PENGUIN_STAGING)

penguin-38: penguin-staging ## Build the penguin image for Python 3.8
	docker build --build-arg PYTHON_VERSION=3.8.17 --build-arg TFX_VERSION=1.9.1 -t penguin:3.8 $(PENGUIN_STAGING)

penguin-39: penguin-staging ## Build the penguin image for Python 3.9
	docker build --build-arg PYTHON_VERSION=3.9.13 --build-arg TFX_VERSION=1.9.1 -t penguin:3.9 $(PENGUIN_STAGING)

penguin-310: penguin-staging ## Build the penguin image for Python 3.10
	docker build --build-arg PYTHON_VERSION=3.10.12 --build-arg TFX_VERSION=1.15.1 -t penguin:3.10 $(PENGUIN_STAGING)

integration-test: docker-build penguin-310 ## Run the integration test
	$(eval TMP := $(shell mktemp -d))
	echo '{"rootLocation": "pipeline_root", "servingLocation": "serving", "name": "test", "image": "test-pipeline", "tfxComponents": "pipeline.create_components"}' > $(TMP)/pipeline_config.json
	cp resources/compile.sh $(TMP)
	docker run -v $(TMP):/shared kfp-operator-tfx-compiler tfx /shared

	#docker run -v $(TMP):/shared --entrypoint /shared/compile.sh penguin:3.8 tfx --output_file /tmp/pipeline.yaml --pipeline_config /shared/pipeline_config.json --execution_mode v2
	#docker run -v $(TMP):/shared --entrypoint /shared/compile.sh penguin:3.9 tfx --output_file /tmp/pipeline.yaml --pipeline_config /shared/pipeline_config.json --execution_mode v2
	docker run -v $(TMP):/shared --entrypoint /shared/compile.sh penguin:3.10 tfx --output_file /tmp/pipeline.yaml --pipeline_config /shared/pipeline_config.json --execution_mode v2
