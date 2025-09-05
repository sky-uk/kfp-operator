include ../../common.mk
include ../../docker-targets.mk
include ../../help.mk

IMG := kfp-operator-kfp-provider-service

##@ Development

unit-test: ## Unit-test the provider-service
	go test ./... -tags=unit

test: unit-test ## Run all tests for the provider-service

##@ Build

build: ## Build the provider-service
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/kfp-provider-service cmd/main.go

