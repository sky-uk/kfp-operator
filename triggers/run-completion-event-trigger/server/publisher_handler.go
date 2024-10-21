package run_completion_event_trigger

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type PublisherHandler interface {
	Publish(runCompletionEvent common.RunCompletionEvent) (*MarshallingError, *ConnectionError)
}

type MarshallingError struct {
	Error error
}

type ConnectionError struct {
	Error error
}

type NatsPublisher struct {
	NatsConn *nats.Conn
	Subject  string
}

func (nc NatsPublisher) Publish(runCompletionEvent common.RunCompletionEvent) (*MarshallingError, *ConnectionError) {
	eventData, err := json.Marshal(runCompletionEvent)
	if err != nil {
		return &MarshallingError{err}, nil
	}
	if err := nc.NatsConn.Publish(nc.Subject, eventData); err != nil {
		return nil, &ConnectionError{err}
	}
	return nil, nil
}
