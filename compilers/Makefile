include ../help.mk

##@ Development

test-all: ## Test all compilers
	$(MAKE) -C tfx test
	$(MAKE) -C kfp-sdk test

integration-test-all: ## Integration test for all compilers
	$(MAKE) -C tfx integration-test
	$(MAKE) -C kfp-sdk integration-test

##@ Build

build-all: ## Build all compilers
	$(MAKE) -C tfx build
	$(MAKE) -C kfp-sdk build

##@ Containers

docker-build-tfx:
	$(MAKE) -C tfx docker-build

docker-build-kfp-sdk:
	$(MAKE) -C kfp-sdk docker-build

docker-build-all: docker-build-tfx docker-build-kfp-sdk ## Build all compiler docker images

docker-push-tfx:
	$(MAKE) -C tfx docker-push

docker-push-kfp-sdk:
	$(MAKE) -C kfp-sdk docker-push

docker-push-all: docker-push-tfx docker-push-kfp-sdk ## Push all compiler docker images
