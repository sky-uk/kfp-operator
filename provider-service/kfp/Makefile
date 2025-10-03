include ../../common.mk
include ../../docker-targets.mk
include ../../help.mk
include ./get-proto.mk

IMG := kfp-operator-kfp-provider-service

##@ Development

unit-test: ## Unit-test the provider-service
	go test ./... -tags=unit

test: unit-test ## Run all tests for the provider-service

##@ Build
generate: protoc-gen-go ## Generate service definitions from protobuf
	$(call get-proto,github.com/google/ml-metadata,v1.17.0)
	protoc --go_out=. --go-grpc_out=. \
		-I $(PROTOPATH)/github.com/google/ml-metadata@v1.17.0/ \
		--go_opt=Mml_metadata/proto/metadata_store_service.proto=/internal/client/ml_metadata \
		--go_opt=Mml_metadata/proto/metadata_store.proto=/internal/client/ml_metadata \
		--go_opt=Mml_metadata/proto/metadata_source.proto=/internal/client/ml_metadata \
		--go-grpc_opt=Mml_metadata/proto/metadata_store_service.proto=/internal/client/ml_metadata \
		--go-grpc_opt=Mml_metadata/proto/metadata_store.proto=/internal/client/ml_metadata \
		ml_metadata/proto/metadata_store_service.proto \
		ml_metadata/proto/metadata_store.proto


build: ## Build the provider-service
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/kfp-provider-service cmd/main.go
