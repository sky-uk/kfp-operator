package logging

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

type loggerKey struct{}

func NewLogger() (*logr.Logger, error) {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	logger := zapr.NewLogger(zapLogger.Named("model-update-eventsource"))
	return &logger, nil
}

func WithLogger(ctx context.Context, logger *logr.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func FromContext(ctx context.Context) *logr.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*logr.Logger); ok {
		return logger
	}

	discardingLogger := logr.Discard()
	return &discardingLogger
}
