include ../help.mk

##@ Development

test-all: ## Test all compilers
	$(MAKE) -C tfx test
	$(MAKE) -C kfpsdk test

integration-test-all: ## Integration test for all compilers
	$(MAKE) -C tfx integration-test
	$(MAKE) -C kfpsdk integration-test

##@ Build

build-all: ## Build all compilers
	$(MAKE) -C tfx build
	$(MAKE) -C kfpsdk build

##@ Containers

docker-build-all: ## Build all compiler docker images
	$(MAKE) -C tfx docker-build
	$(MAKE) -C kfpsdk docker-build

docker-push-all: ## Push all compiler docker images
	$(MAKE) -C tfx docker-push
	$(MAKE) -C kfpsdk docker-push
