##@ Development
unit-test:
	go test ./... -tags=unit

decoupled-test:
	go test ./... -tags=decoupled

test: unit-test decoupled-test integration-test

integration-test:
	docker compose up -d
	$(MAKE) wait-for-pubsub
	go test ./... -v -tags=integration
	docker compose down

wait-for-pubsub:
	@echo "Checking if Pub/Sub Emulator is healthy..."
	@while [[ "$$(docker inspect --format='{{.State.Health.Status}}' pubsub 2>/dev/null)" != "healthy" ]]; do \
		echo "Waiting for Pub/Sub Emulator to be healthy..."; \
		sleep 5; \
	done
	@echo "Pub/Sub Emulator is healthy!"

