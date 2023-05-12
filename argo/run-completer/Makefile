include ../../common.mk

IMG := kfp-operator-run-completer

all: build

##@ Development

unit-test:
	go test ./... -tags=unit

decoupled-test:
	$(call envtest-run,go test ./... -tags=decoupled -coverprofile cover.out)

test: unit-test decoupled-test

##@ Build

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/run-completer ./cmd

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

##@ Containers

include ../../docker-targets.mk
