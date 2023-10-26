include ../../common.mk

IMG := kfp-operator-kfp-provider

all: build

##@ Development

base-test:
	go test ./base/... -tags=unit

test: base-test
	$(MAKE) -C kfp test
	$(MAKE) -C vai test

##@ Build

generate: protoc-gen-go
	$(MAKE) -C kfp generate

build: generate
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
