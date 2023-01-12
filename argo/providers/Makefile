include ../../common.mk
include get-proto.mk

IMG := kfp-operator-kfp-provider

all: build

##@ Development

base-test:
	go test ./base/... -tags=unit

test: base-test
	$(MAKE) -f kfp.mk test
	$(MAKE) -f vai.mk test

##@ Build

generate: protoc-gen-go
	$(call get-proto,github.com/argoproj/argo-events,v1.5.4)
	protoc --go_out=. --go-grpc_out=base/generic \
		-I $(PROTOPATH)/github.com/argoproj/argo-events@v1.5.4/ \
		--go_opt=Meventsources/sources/generic/generic.proto=/base/generic \
		--go-grpc_opt=module=eventsources/sources/generic \
		eventsources/sources/generic/generic.proto

	$(MAKE) -f kfp.mk generate

build: generate
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
