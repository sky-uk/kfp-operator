generate:
	protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           eventsources/model_update/server/model_update_eventsource.proto

build: generate
	go build -o bin/model-update-eventsource ./eventsources/model_update/cmd/main.go

run-model-update-eventsource: build fmt vet
	go run ./eventsources/model_update/cmd/main.go

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...
