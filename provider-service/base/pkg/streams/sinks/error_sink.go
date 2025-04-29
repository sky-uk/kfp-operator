package sinks

import (
	"context"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type ErrorSink struct {
	in chan error
}

func NewErrorSink(ctx context.Context, inChan chan error) *ErrorSink {
	errorSink := &ErrorSink{in: inChan}

	go errorSink.Log(ctx)

	return errorSink
}

func (es ErrorSink) In() chan<- error {
	return es.in
}

func (es ErrorSink) Log(ctx context.Context) {
	logger := common.LoggerFromContext(ctx)
	for err := range es.in {
		logger.Error(err, "failed to handle event")
	}
}
