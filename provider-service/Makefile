include ../help.mk

##@ General

build: ## Build all providers
	$(MAKE) -C vai build

##@ Development

unit-test: ## Unit-test all providers
	$(MAKE) -C base unit-test
	$(MAKE) -C vai unit-test
	$(MAKE) -C kfp unit-test

decoupled-test: ## Run all decoupled tests
	$(MAKE) -C base decoupled-test

integration-test: ## Run all integration tests
	$(MAKE) -C base integration-test

test: ## Run all tests for all providers
	$(MAKE) -C base test
	$(MAKE) -C vai test
	$(MAKE) -C kfp test

##@ Containers

docker-build: ## Build all provider docker images
	$(MAKE) -C vai docker-build
	$(MAKE) -C kfp docker-build


docker-push: ## Publish all provider docker images
	$(MAKE) -C vai docker-push
	$(MAKE) -C kfp docker-push


