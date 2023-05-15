include ../../../common.mk

IMG := kfp-operator-vai-provider

##@ Development

unit-test:
	go test ./... -tags=unit

test: unit-test

##@ Build

MOCKGEN := $(PROJECT_DIR)/bin/mockgen
mockgen: ## Download mockgen locally if necessary.
	$(call go-install,$(PROJECT_DIR)/bin/mockgen,github.com/golang/mock/mockgen@v1.6.0)

generate: mockgen
	$(MOCKGEN) -destination mock_pipeline_client.go -package=vai github.com/sky-uk/kfp-operator/argo/providers/vai PipelineJobClient

build: generate
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/provider ./cmd

##@ Containers

include ../../../docker-targets.mk
