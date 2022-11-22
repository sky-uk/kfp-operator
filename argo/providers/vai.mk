include ../../common.mk

IMG := kfp-operator-vai-provider

##@ Development

test:
	go test ./... -tags=unit

##@ Build

MOCKGEN := $(PROJECT_DIR)/bin/mockgen
mockgen: ## Download mockgen locally if necessary.
	$(call go-install,$(PROJECT_DIR)/bin/mockgen,github.com/golang/mock/mockgen@v1.6.0)

generate: mockgen
	$(MOCKGEN) -destination vai/mock_pipeline_client.go -package=vai github.com/sky-uk/kfp-operator/providers/vai PipelineJobClient

build: generate
	go build -o bin/provider ./vai/cmd

##@ Containers

DOCKER_BUILD_EXTRA_PARAMS=-f vai/Dockerfile
include ../../docker-targets.mk
