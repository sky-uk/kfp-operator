package main

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
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

	// TODO: load in the config
	cfg := config.Server{
		Host: "localhost",
		Port: 8080,
	}
	if err = server.Start(ctx, cfg, provider); err != nil {
		panic(err)
	}

}
