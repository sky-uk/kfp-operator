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

func (ls ErrorSink) In() chan<- error {
	return ls.in
}

func (ls ErrorSink) Log() {
	logger := common.LoggerFromContext(ls.context)
	for err := range ls.in {
		logger.Error(err, "Failed to handle event")
	}
}
