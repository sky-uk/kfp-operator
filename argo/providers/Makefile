IMG := kfp-operator-kfp-provider

all: build

##@ Development

base-test:
	go test ./base/... -tags=unit

test: base-test
	$(MAKE) -f kfp.mk test
	$(MAKE) -f vai.mk test

##@ Build

build:
	$(MAKE) -f kfp.mk build
	$(MAKE) -f vai.mk build

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

##@ Containers

docker-build:
	$(MAKE) -f kfp.mk docker-build
	$(MAKE) -f vai.mk docker-build

docker-push:
	$(MAKE) -f kfp.mk docker-push
	$(MAKE) -f vai.mk docker-push
