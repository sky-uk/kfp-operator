package publisher

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
	"github.com/sky-uk/kfp-operator/pkg/common"
)

type natsConn interface {
	Publish(subject string, data []byte) error
	IsConnected() bool
}

type NatsPublisher struct {
	NatsConn natsConn
	Subject  string
}

func NewNatsPublisher(ctx context.Context, nc *nats.Conn, subject string) *NatsPublisher {
	logger := logr.FromContextOrDiscard(ctx)

	logger.Info("New nats publisher:", "Subject", subject, "Server", nc.ConnectedUrl())
	return &NatsPublisher{
		NatsConn: nc,
		Subject:  subject,
	}
}

func (nc *NatsPublisher) Publish(_ context.Context, runCompletionEvent common.RunCompletionEvent) error {
	dataWrapper := DataWrapper{Data: runCompletionEvent}
	eventData, err := json.Marshal(dataWrapper)
	if err != nil {
		return &MarshallingError{err.Error()}
	}
	if err := nc.NatsConn.Publish(nc.Subject, eventData); err != nil {
		return &ConnectionError{err.Error()}
	}
	return nil
}

func (nc *NatsPublisher) Name() string {
	return "nats-publisher"
}

func (nc *NatsPublisher) IsHealthy() bool {
	return nc.NatsConn.IsConnected()
}

// Close closes the NATS connection
func (nc *NatsPublisher) Close() {
	if conn, ok := nc.NatsConn.(*nats.Conn); ok {
		conn.Close()
	}
}
