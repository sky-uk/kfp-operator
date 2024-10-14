package nats_event_trigger

import (
	"github.com/nats-io/nats.go"
	"log"
)

type PublisherHandler interface {
	Publish(data []byte) error
}

type NatsPublisher struct {
	NatsConn *nats.Conn
	Subject  string
}

func (nc NatsPublisher) Publish(data []byte) error {
	log.Printf("Publish called")
	if err := nc.NatsConn.Publish(nc.Subject, data); err != nil {
		return err
	}
	return nil
}
