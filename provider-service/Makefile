include ../help.mk

##@ General

build: ## Build all providers
	$(MAKE) -C vai build
	$(MAKE) -C kfp build

##@ Development

unit-test: ## Unit-test all providers
	$(MAKE) -C base unit-test
	$(MAKE) -C vai unit-test
	$(MAKE) -C kfp unit-test

decoupled-test: ## Run all decoupled tests
	$(MAKE) -C base decoupled-test
	$(MAKE) -C kfp decoupled-test

integration-test: ## Run all integration tests
	$(MAKE) -C base integration-test

test: ## Run all tests for all providers
	$(MAKE) -C base test
	$(MAKE) -C vai test
	$(MAKE) -C kfp test

##@ Containers

docker-build-vai:
	$(MAKE) -C vai docker-build

docker-build-kfp:
	$(MAKE) -C kfp docker-build

docker-build: docker-build-vai docker-build-kfp ## Build all provider docker images

docker-push-vai:
	$(MAKE) -C vai docker-push

docker-push-kfp:
	$(MAKE) -C kfp docker-push

docker-push: docker-push-vai docker-push-kfp ## Publish all provider docker images
