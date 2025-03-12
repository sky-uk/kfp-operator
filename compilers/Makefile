include ../help.mk

##@ Development

test-all: ## Test all compilers
	$(MAKE) -C tfx test

integration-test-all: ## Integration test for all compilers
	$(MAKE) -C tfx integration-test

##@ Build

build-all: ## Build all compilers
	$(MAKE) -C tfx build

##@ Containers

docker-build-all: ## Build all compiler docker images
	$(MAKE) -C tfx docker-build

docker-push-all: ## Push all compiler docker images
	$(MAKE) -C tfx docker-push
