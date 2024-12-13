package sinks

import (
	"context"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type ErrorSink struct {
	context context.Context
	in      chan error
}

func NewErrorSink(ctx context.Context, inChan chan error) *ErrorSink {
	errorSink := &ErrorSink{context: ctx, in: inChan}

	go errorSink.Log()

	return errorSink
}

func (es ErrorSink) In() chan<- error {
	return es.in
}

func (es ErrorSink) Log() {
	logger := common.LoggerFromContext(es.context)
	for err := range es.in {
		logger.Error(err, "failed to handle event")
	}
}
