package publisher

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/sky-uk/kfp-operator/argo/common"
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

type NatsPublisher struct {
	NatsConn *nats.Conn
	Subject  string
}

func (nc NatsPublisher) Publish(runCompletionEvent common.RunCompletionEvent) error {
	eventData, err := json.Marshal(runCompletionEvent)
	if err != nil {
		return &MarshallingError{err.Error()}
	}
	if err := nc.NatsConn.Publish(nc.Subject, eventData); err != nil {
		return &ConnectionError{err.Error()}
	}
	return nil
}
