include ../help.mk

##@ Development

test-all: ## Test all compilers
	#$(MAKE) -C tfx test
	$(MAKE) -C kfp-sdk test

integration-test-all: ## Integration test for all compilers
	$(MAKE) -C tfx integration-test
	$(MAKE) -C kfp-sdk integration-test

##@ Build

build-all: ## Build all compilers
	$(MAKE) -C tfx build
	$(MAKE) -C kfp-sdk build

##@ Containers

docker-build-all: ## Build all compiler docker images
	$(MAKE) -C tfx docker-build
	$(MAKE) -C kfp-sdk docker-build

docker-push-all: ## Push all compiler docker images
	$(MAKE) -C tfx docker-push
	$(MAKE) -C kfp-sdk docker-push
