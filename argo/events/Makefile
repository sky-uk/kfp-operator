##@ Build

include ../../go-get.mk

PROTOPATH := $(HOME)/.proto
IMG := kfp-operator-argo-events-run-completion-eventsource-server

define get-proto
	mkdir -p $(PROTOPATH) && \
	cd $(PROTOPATH); \
	[ -d "$(1)@$(2)" ] || { \
		git clone --no-checkout --branch $(2) --single-branch https://$(1) $(1)@$(2) && \
		cd $(1)@$(2) && \
		git sparse-checkout set *.proto && \
		git checkout tags/$(2) -b master; \
	}
endef

protoc_gen_go: ## Download protoc-gen-go locally if necessary.
	$(call go-get-tool,$(PROJECT_DIR)/bin/protoc-gen-go,google.golang.org/protobuf/cmd/protoc-gen-go@v1.26)
	$(call go-get-tool,$(PROJECT_DIR)/bin/protoc-gen-go-grpc,google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1)

mockgen: ## Download mockgen locally if necessary.
	$(call go-get-tool,$(PROJECT_DIR)/bin/mockgen,github.com/golang/mock/mockgen@v1.6.0)

generate: protoc_gen_go mockgen ## Generate service definitions from protobuf
	$(call get-proto,github.com/google/ml-metadata,v1.5.0)
	protoc --go_out=. --go-grpc_out=ml_metadata \
    		-I $(PROTOPATH)/github.com/google/ml-metadata@v1.5.0/ \
    		--go_opt=Mml_metadata/proto/metadata_store_service.proto=/ml_metadata \
    		--go_opt=Mml_metadata/proto/metadata_store.proto=/ml_metadata \
    		--go_opt=Mml_metadata/proto/metadata_source.proto=/ml_metadata \
    		--go-grpc_opt=module=ml_metadata/proto \
    		ml_metadata/proto/metadata_store_service.proto \
    		ml_metadata/proto/metadata_store.proto

	$(call get-proto,github.com/argoproj/argo-events,v1.5.4)
	protoc --go_out=. --go-grpc_out=eventsources/generic \
		-I $(PROTOPATH)/github.com/argoproj/argo-events@v1.5.4/ \
		--go_opt=Meventsources/sources/generic/generic.proto=/eventsources/generic \
		--go-grpc_opt=module=eventsources/sources/generic \
	  	eventsources/sources/generic/generic.proto

	mockgen -destination ml_metadata/metadata_store_service_grpc_mock.go -package=ml_metadata -source=ml_metadata/metadata_store_service_grpc.pb.go

build: generate
	go build -o bin/run-completion-eventsource ./eventsources/run_completion/cmd/main.go

run-run-completion-eventsource: build fmt vet
	go run ./eventsources/run_completion/cmd/main.go -mlmd-url=$(MLMD_URL)

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

##@ Development

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
decoupled-test: ## Run decoupled acceptance tests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.8.3/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -tags=decoupled -coverprofile cover.out

unit-test: ## Run unit tests
	go test ./... -tags=unit

test: generate fmt vet unit-test

##@ Containers

include ../../docker-targets.mk
