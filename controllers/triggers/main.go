package main

import (
	"log"

	"github.com/sky-uk/kfp-operator/controllers/triggers/nats_event_trigger"
)

func main() {
	err := nats_event_trigger.Start()
	if err != nil {
		log.Fatal("Failed to start server: %w", err)
	}
}
