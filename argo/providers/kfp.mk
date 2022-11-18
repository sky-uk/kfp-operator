include ../../go-get.mk
include get-proto.mk
IMG := kfp-operator-kfp-provider

all: build

##@ Build

MOCKGEN := $(PROJECT_DIR)/bin/mockgen
mockgen: ## Download mockgen locally if necessary.
	$(call go-install,$(PROJECT_DIR)/bin/mockgen,github.com/golang/mock/mockgen@v1.6.0)

generate: protoc_gen_go mockgen ## Generate service definitions from protobuf
	$(call get-proto,github.com/google/ml-metadata,v1.5.0)
	protoc --go_out=. --go-grpc_out=kfp/ml_metadata \
		-I $(PROTOPATH)/github.com/google/ml-metadata@v1.5.0/ \
		--go_opt=Mml_metadata/proto/metadata_store_service.proto=/kfp/ml_metadata \
		--go_opt=Mml_metadata/proto/metadata_store.proto=/kfp/ml_metadata \
		--go_opt=Mml_metadata/proto/metadata_source.proto=/kfp/ml_metadata \
		--go-grpc_opt=module=ml_metadata/proto \
		ml_metadata/proto/metadata_store_service.proto \
		ml_metadata/proto/metadata_store.proto

	$(MOCKGEN) -destination kfp/ml_metadata/metadata_store_service_grpc_mock.go -package=ml_metadata -source=kfp/ml_metadata/metadata_store_service_grpc.pb.go
	$(MOCKGEN) -destination kfp/run_service_grpc_mock.go -package=kfp github.com/kubeflow/pipelines/backend/api/go_client RunServiceClient

##@ Development

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
decoupled-test: ## Run decoupled acceptance tests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.8.3/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -tags=decoupled -coverprofile cover.out

unit-test:
	go test ./... -tags=unit

test-python: build-sdk
	poetry run pytest

test: test-python unit-test decoupled-test

##@ Build

build-sdk:
	pip install poetry-dynamic-versioning --quiet
	poetry install
	poetry build

build-go: generate
	go build -o bin/provider ./kfp/cmd

build: build-sdk build-go

##@ Containers

WHEEL_VERSION=$(shell poetry version | cut -d ' ' -f 2)
DOCKER_BUILD_EXTRA_PARAMS=-f kfp/Dockerfile --build-arg WHEEL_VERSION=${WHEEL_VERSION}
include ../../docker-targets.mk
