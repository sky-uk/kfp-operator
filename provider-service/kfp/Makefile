include ../../common.mk
include ../../docker-targets.mk
include ./get-proto.mk

IMG := kfp-operator-kfp-provider-service

all: build

##@ Build

generate: protoc-gen-go ## Generate service definitions from protobuf
	$(call get-proto,github.com/google/ml-metadata,v1.5.0)
	protoc --go_out=. --go-grpc_out=internal/client/ml_metadata \
		-I $(PROTOPATH)/github.com/google/ml-metadata@v1.5.0/ \
		--go_opt=Mml_metadata/proto/metadata_store_service.proto=/internal/client/ml_metadata \
		--go_opt=Mml_metadata/proto/metadata_store.proto=/internal/client/ml_metadata \
		--go_opt=Mml_metadata/proto/metadata_source.proto=/internal/client/ml_metadata \
		--go-grpc_opt=module=ml_metadata/proto \
		ml_metadata/proto/metadata_store_service.proto \
		ml_metadata/proto/metadata_store.proto

##@ Development

decoupled-test: ## Run decoupled acceptance tests
	$(call envtest-run,go test ./... -tags=decoupled -coverprofile cover.out)

unit-test:
	go test ./... -tags=unit

test: unit-test decoupled-test

##@ Build

build: generate
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/kfp-provider-service cmd/main.go

##@ Containers

DOCKER_BUILD_EXTRA_PARAMS=--build-context base-source=".."
