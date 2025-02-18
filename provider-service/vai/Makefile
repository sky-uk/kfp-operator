include ../../common.mk
include ../../docker-targets.mk
include ../../help.mk

IMG := kfp-operator-vai-provider-service
DOCKER_BUILD_EXTRA_PARAMS=--build-context base-source=".."

##@ Development

unit-test: ## Unit-test the provider-service
	go test ./... -tags=unit

test: unit-test ## Run all tests for the provider-service

##@ Build

build: ## Build the provider-service
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/vai-provider-service cmd/main.go

