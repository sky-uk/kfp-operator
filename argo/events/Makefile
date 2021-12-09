generate:
	##protoc --go_out=. --go-grpc_out=. eventsources/api/eventing.proto
	protoc --go_out=eventsources --go-grpc_out=eventsources/generic \
		-I $(GOPATH)/pkg/mod/github.com/argoproj/argo-events@v1.15.0/eventsources/sources/generic/ \
		--go_opt=Mgeneric.proto=/generic \
		generic.proto
	## protoc -I mlmd --go_out=. --go-grpc_out=. mlmd/metadata_store_service.proto

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
decoupled-test: ## Run decoupled acceptance tests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.8.3/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -tags=decoupled -parallel 1 -coverprofile cover.out

unit-test: ## Run unit tests
	go test ./... -tags=unit

test: generate fmt vet unit-test

build: generate
	go build -o bin/model-update-eventsource ./eventsources/model_update/cmd/main.go

run-model-update-eventsource: build fmt vet
	go run ./eventsources/model_update/cmd/main.go

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...
