package publisher

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/sky-uk/kfp-operator/pkg/common"
)

type PublisherHandler interface {
	Publish(runCompletionEvent common.RunCompletionEvent) error
}

type MarshallingError struct {
	Message string
}

func (e *MarshallingError) Error() string {
	return e.Message
}

type ConnectionError struct {
	Message string
}

func (e *ConnectionError) Error() string {
	return e.Message
}

type natsConn interface {
	Publish(subject string, data []byte) error
	IsConnected() bool
}

type NatsPublisher struct {
	NatsConn natsConn
	Subject  string
}

type DataWrapper struct {
	Data common.RunCompletionEvent `json:"data"`
}

func NewNatsPublisher(ctx context.Context, nc *nats.Conn, subject string) *NatsPublisher {
	logger := common.LoggerFromContext(ctx)

	logger.Info("New nats publisher:", "Subject", subject, "Server", nc.ConnectedUrl())
	return &NatsPublisher{
		NatsConn: nc,
		Subject:  subject,
	}
}

func (nc *NatsPublisher) Publish(runCompletionEvent common.RunCompletionEvent) error {
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
