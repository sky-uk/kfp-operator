package main

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/sky-uk/kfp-operator/argo/common"
	baseConfig "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server"
	"github.com/sky-uk/kfp-operator/provider-service/stub/internal/provider"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger, err := common.NewLogger(zapcore.InfoLevel)
	if err != nil {
		panic(err)
	}

	ctx := logr.NewContext(context.Background(), logger)
	provider := provider.New(logger)
	cfg, err := baseConfig.LoadConfig(
		baseConfig.Config{
			Server: baseConfig.Server{
				Host: "0.0.0.0",
				Port: 8080,
			},
			Metrics: baseConfig.MetricsConfig{
				Port: 8181,
			},
		},
	)
	if err != nil {
		panic(err)
	}

	if err = server.Start(ctx, *cfg, provider); err != nil {
		panic(err)
	}

	<-ctx.Done()
	logger.Info("Main context is cancelled. Terminating application...")
}
