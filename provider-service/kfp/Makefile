include ../../common.mk
include ../../docker-targets.mk
include ./get-proto.mk

IMG := kfp-operator-kfp-provider-service

all: build

##@ Build

MOCKGEN := $(PROJECT_DIR)/bin/mockgen
mockgen: ## Download mockgen locally if necessary.
	$(call go-install,$(PROJECT_DIR)/bin/mockgen,github.com/golang/mock/mockgen@v1.6.0)

generate: protoc-gen-go mockgen ## Generate service definitions from protobuf
	$(call get-proto,github.com/google/ml-metadata,v1.5.0)
	protoc --go_out=. --go-grpc_out=internal/ml_metadata \
		-I $(PROTOPATH)/github.com/google/ml-metadata@v1.5.0/ \
		--go_opt=Mml_metadata/proto/metadata_store_service.proto=/internal/ml_metadata \
		--go_opt=Mml_metadata/proto/metadata_store.proto=/internal/ml_metadata \
		--go_opt=Mml_metadata/proto/metadata_source.proto=/internal/ml_metadata \
		--go-grpc_opt=module=ml_metadata/proto \
		ml_metadata/proto/metadata_store_service.proto \
		ml_metadata/proto/metadata_store.proto

	$(MOCKGEN) -destination internal/ml_metadata/metadata_store_service_grpc_mock.go -package=ml_metadata -source=internal/ml_metadata/metadata_store_service_grpc.pb.go
	$(MOCKGEN) -destination internal/run_service_grpc_mock.go -package=internal github.com/kubeflow/pipelines/backend/api/go_client RunServiceClient
	$(MOCKGEN) -destination internal/job_service_grpc_mock.go -package=internal github.com/kubeflow/pipelines/backend/api/go_client JobServiceClient

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
