include ../../docker-targets.mk
include ../../help.mk

IMG := kfp-operator-run-completion-event-trigger

##@ Build

build: ## Build run-completion-event-trigger
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/run-completion-event-trigger ./cmd/main.go

##@ Development
unit-test: ## Run unit tests
	go test ./... -tags=unit

functional-test: docker-build ## Run all functional tests
	docker compose up -d
	go test ./... -tags=functional
	docker compose down

test: unit-test functional-test ## Run all tests
