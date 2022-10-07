IMG := kfp-operator-argo-kfp-sdk

all: build

##@ Development

test: build-sdk
	poetry run pytest

##@ Build

build-sdk:
	pip install poetry-dynamic-versioning --quiet
	poetry install
	poetry build

build-go:
	go build -o bin/provider ./kfp/main.go

build: build-sdk build-go

run: build fmt vet
	go run ./cmd/main.go

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...
##@ Containers

WHEEL_VERSION=$(shell poetry version | cut -d ' ' -f 2)
DOCKER_BUILD_EXTRA_PARAMS=-f kfp/Dockerfile --build-arg WHEEL_VERSION=${WHEEL_VERSION}
include ../../docker-targets.mk
