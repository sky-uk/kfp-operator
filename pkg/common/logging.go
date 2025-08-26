package common

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func LoggerFromContext(ctx context.Context) logr.Logger {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		return logr.Discard()
	}

	return logger
}

func NewLogger(logLevel zapcore.Level) (logr.Logger, error) {
	config := zap.NewProductionConfig()
	config.Level.SetLevel(logLevel)
	zapLogger, err := config.Build()
	if err != nil {
		return logr.Discard(), err
	}

	return zapr.NewLogger(zapLogger.Named("main")), nil
}
