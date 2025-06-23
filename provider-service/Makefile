include ../help.mk

##@ General

build: ## Build all providers
	$(MAKE) -C vai build

##@ Development

unit-test: ## Unit-test all providers
	$(MAKE) -C base unit-test
	$(MAKE) -C vai unit-test

decoupled-test: ## Run all decoupled tests
	$(MAKE) -C base decoupled-test

test: ## Run all tests for all providers
	$(MAKE) -C base test
	$(MAKE) -C vai test

##@ Containers

docker-build: ## Build all provider docker images
	$(MAKE) -C vai docker-build

docker-push: ## Publish all provider docker images
	$(MAKE) -C vai docker-push

