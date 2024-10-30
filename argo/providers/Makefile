include ../../common.mk

all: build

##@ Development

base-test:
	go test ./base/... -tags=unit

test: base-test
	$(MAKE) -C vai test

build:
	$(MAKE) -C kfp build
	$(MAKE) -C vai build

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

##@ Containers

docker-build:
	$(MAKE) -C kfp docker-build
	$(MAKE) -C vai docker-build

docker-push:
	$(MAKE) -C kfp docker-push
	$(MAKE) -C vai docker-push
