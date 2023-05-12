include ../../../common.mk
include ../get-proto.mk
IMG := kfp-operator-kfp-provider

all: build

##@ Build

MOCKGEN := $(PROJECT_DIR)/bin/mockgen
mockgen: ## Download mockgen locally if necessary.
	$(call go-install,$(PROJECT_DIR)/bin/mockgen,github.com/golang/mock/mockgen@v1.6.0)

generate: protoc-gen-go mockgen ## Generate service definitions from protobuf
	$(call get-proto,github.com/google/ml-metadata,v1.5.0)
	protoc --go_out=. --go-grpc_out=ml_metadata \
		-I $(PROTOPATH)/github.com/google/ml-metadata@v1.5.0/ \
		--go_opt=Mml_metadata/proto/metadata_store_service.proto=/ml_metadata \
		--go_opt=Mml_metadata/proto/metadata_store.proto=/ml_metadata \
		--go_opt=Mml_metadata/proto/metadata_source.proto=/ml_metadata \
		--go-grpc_opt=module=ml_metadata/proto \
		ml_metadata/proto/metadata_store_service.proto \
		ml_metadata/proto/metadata_store.proto

	$(MOCKGEN) -destination ml_metadata/metadata_store_service_grpc_mock.go -package=ml_metadata -source=ml_metadata/metadata_store_service_grpc.pb.go
	$(MOCKGEN) -destination run_service_grpc_mock.go -package=kfp github.com/kubeflow/pipelines/backend/api/go_client RunServiceClient
	$(MOCKGEN) -destination job_service_grpc_mock.go -package=kfp github.com/kubeflow/pipelines/backend/api/go_client JobServiceClient

##@ Development

decoupled-test: ## Run decoupled acceptance tests
	$(call envtest-run,go test ./... -tags=decoupled -coverprofile cover.out)

unit-test:
	go test ./... -tags=unit

test: unit-test decoupled-test

##@ Build

build: generate
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/provider ./cmd

##@ Containers

include ../../../docker-targets.mk
