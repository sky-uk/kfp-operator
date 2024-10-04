package main

import (
	"github.com/sky-uk/kfp-operator/controllers/triggers/nats_event_trigger"
)

func main() {
	nats_event_trigger.Start()
}
