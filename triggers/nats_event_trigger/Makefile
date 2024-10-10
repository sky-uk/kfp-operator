include ../../docker-targets.mk

IMG := kfp-operator-nats-event-trigger

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/nats_event_trigger ./main.go
