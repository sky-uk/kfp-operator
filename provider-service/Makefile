include ../help.mk

##@ General

build: ## Build all providers
	$(MAKE) -C kfp build
	$(MAKE) -C vai build

##@ Development

unit-test: ## Unit-test all providers
	$(MAKE) -C base unit-test
	$(MAKE) -C kfp unit-test
	$(MAKE) -C vai unit-test

test: ## Run all tests for all providers
	$(MAKE) -C base test
	$(MAKE) -C kfp test
	$(MAKE) -C vai test

##@ Containers

docker-build: ## Build all provider docker images
	$(MAKE) -C kfp docker-build
	$(MAKE) -C vai docker-build

docker-push: ## Publish all provider docker images
	$(MAKE) -C kfp docker-push
	$(MAKE) -C vai docker-push


