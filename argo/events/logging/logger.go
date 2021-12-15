package logging

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerKey struct{}

func NewLogger(logLevel zapcore.Level) (logr.Logger, error) {
	config := zap.NewProductionConfig()
	config.Level.SetLevel(logLevel)
	zapLogger, err := config.Build()
	if err != nil {
		return logr.Discard(), err
	}

	return zapr.NewLogger(zapLogger.Named("model-update-eventsource")), nil
}

func WithLogger(ctx context.Context, logger logr.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func FromContext(ctx context.Context) logr.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(logr.Logger); ok {
		return logger
	}

	return logr.Discard()
}
