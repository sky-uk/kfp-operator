include ../../help.mk

##@ Development
generate: ## Run all golang code generation tools
	go generate ./...

unit-test: generate ## Unit-test the provider-service
	go test ./... -tags=unit

decoupled-test: generate ## Run decoupled acceptance tests
	go test ./... -tags=decoupled

test: unit-test decoupled-test integration-test ## Run all tests

integration-test: generate ## Run integration tests
	docker compose up -d
	$(MAKE) wait-for-pubsub
	go test ./... -v -tags=integration
	docker compose down

wait-for-pubsub: ## Wait for Pub/Sub Emulator to be healthy
	@echo "Checking if Pub/Sub Emulator is healthy..."
	@while [[ "$$(docker inspect --format='{{.State.Health.Status}}' pubsub 2>/dev/null)" != "healthy" ]]; do \
		echo "Waiting for Pub/Sub Emulator to be healthy..."; \
		sleep 5; \
	done
	@echo "Pub/Sub Emulator is healthy!"
