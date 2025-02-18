include ../../common.mk
include ../../help.mk

all: build

##@ Build
build: ## Build all providers
	$(MAKE) -C kfp build
	$(MAKE) -C vai build

##@ Development

base-test: ## Unit-test the base provider
	go test ./base/... -tags=unit

test: base-test ## Run all tests for all providers
	$(MAKE) -C kfp test
	$(MAKE) -C vai test

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

##@ Containers

docker-build: ## Build all provider docker images
	$(MAKE) -C kfp docker-build
	$(MAKE) -C vai docker-build

docker-push: ## Publish all provider docker images
	$(MAKE) -C kfp docker-push
	$(MAKE) -C vai docker-push
