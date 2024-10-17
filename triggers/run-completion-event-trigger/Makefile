include ../../docker-targets.mk

IMG := kfp-operator-run-completion-event-trigger

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/run-completion-event-trigger ./main.go

test:
	go test ./... -tags=unit

functional-test: docker-build
	docker compose up -d
	go test ./... -tags=functional
	docker compose down
