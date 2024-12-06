##@ Development
unit-test:
	go test ./... -tags=unit

test: unit-test

integration-test:
	docker compose up -d
	$(MAKE) wait-for-pubsub
	go test ./... -v -tags=integration
	docker compose down

# Wait for the Pub/Sub emulator to become healthy
wait-for-pubsub:
	@echo "Checking if Pub/Sub Emulator is healthy..."
	@while [[ "$$(docker inspect --format='{{.State.Health.Status}}' pubsub 2>/dev/null)" != "healthy" ]]; do \
		echo "Waiting for Pub/Sub Emulator to be healthy..."; \
		sleep 5; \
	done
	@echo "Pub/Sub Emulator is healthy!"

