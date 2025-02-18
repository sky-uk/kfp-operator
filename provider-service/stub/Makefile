include ../../common.mk
include ../../docker-targets.mk
include ../../help.mk

IMG := kfp-operator-stub-provider-service

all: build

##@ Development

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

##@ Build

build: ## Build stub provider service binary
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/stub-provider-service cmd/main.go
