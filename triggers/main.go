package main

import (
	"log"

	server "github.com/sky-uk/kfp-operator/triggers/nats_event_trigger/server"
)

func main() {
	err := server.Start()
	if err != nil {
		log.Fatal("Failed to start server: %w", err)
	}
}
