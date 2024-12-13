include ../../common.mk
include ../../docker-targets.mk

IMG := kfp-operator-vai-provider-service

##@ Development

unit-test:
	go test ./... -tags=unit

test: unit-test

##@ Build

MOCKGEN := $(PROJECT_DIR)/bin/mockgen
mockgen: ## Download mockgen locally if necessary.
	$(call go-install,$(PROJECT_DIR)/bin/mockgen,github.com/golang/mock/mockgen@v1.6.0)

generate: mockgen
	$(MOCKGEN) -destination internal/mock_pipeline_client.go -package=internal github.com/sky-uk/kfp-operator/provider-service/vai/internal PipelineJobClient

build: generate
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/vai-provider-service cmd/main.go

##@ Containers

DOCKER_BUILD_EXTRA_PARAMS=--build-context base-source=".."
