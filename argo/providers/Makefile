include ../../common.mk
include get-proto.mk

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
	$(call get-proto,github.com/argoproj/argo-events,v1.5.4)
	protoc --go_out=. --go-grpc_out=base/generic \
		-I $(PROTOPATH)/github.com/argoproj/argo-events@v1.5.4/ \
		--go_opt=Meventsources/sources/generic/generic.proto=/base/generic \
		--go-grpc_opt=module=eventsources/sources/generic \
		eventsources/sources/generic/generic.proto

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
