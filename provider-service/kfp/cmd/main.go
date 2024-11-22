package main

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/sky-uk/kfp-operator/argo/common"
	configLoader "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	kfp "github.com/sky-uk/kfp-operator/provider-service/kfp/internal"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger, err := common.NewLogger(zapcore.InfoLevel)
	if err != nil {
		panic(err)
	}
	ctx := logr.NewContext(context.Background(), logger)

	config, err := configLoader.LoadConfig(ctx)
	if err != nil {
		logger.Error(err, "Failed loading configuration")
		panic(err)
	}
	kfp.Start(ctx, *config)
}