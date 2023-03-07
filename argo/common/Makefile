include ../../common.mk

all: build

##@ Development

unit-test:
	go test ./... -tags=unit

test: unit-test

##@ Build

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...
