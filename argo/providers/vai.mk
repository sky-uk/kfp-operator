IMG := kfp-operator-argo-vertex-ai-provider

##@ Development

test:
	@echo no tests defined

##@ Build

build:
	go build -o bin/provider ./vai/main.go

run: build fmt vet
	go run ./cmd/main.go

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

##@ Containers

DOCKER_BUILD_EXTRA_PARAMS=-f vai/Dockerfile
include ../../docker-targets.mk
