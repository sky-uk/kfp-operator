include ../../common.mk
include ../../help.mk

all: build

##@ Development

unit-test: ## Run unit test suites
	go test ./... -tags=unit

test: unit-test ## Run all tests

##@ Build

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...
