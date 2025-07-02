package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-logr/logr"
	"github.com/sky-uk/kfp-operator/internal/log"
	baseConfig "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server"
	kfpConfig "github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	kfp "github.com/sky-uk/kfp-operator/provider-service/kfp/internal/provider"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

func main() {
	logger, err := log.NewLogger(zapcore.InfoLevel)
	if err != nil {
		panic(err)
	}

	rootCtx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	ctx := logr.NewContext(rootCtx, logger)

	serviceConfig, err := baseConfig.LoadConfig(
		baseConfig.Config{
			Server: baseConfig.Server{
				Host: "0.0.0.0",
				Port: 8080,
			},
		},
	)
	if err != nil {
		logger.Error(err, "failed to load config")
		panic(err)
	}
	logger.Info(fmt.Sprintf("loaded base config: %+v", serviceConfig))

	kfpProviderConfig, err := baseConfig.LoadConfig(
		kfpConfig.Config{
			Name:                serviceConfig.ProviderName,
			PipelineRootStorage: serviceConfig.PipelineRootStorage,
		},
	)
	if err != nil {
		logger.Error(err, "failed to load provider config", "provider", serviceConfig.ProviderName, "namespace", serviceConfig.Pod.Namespace)
		panic(err)
	}
	logger.Info(fmt.Sprintf("loaded provider config: %+v", kfpProviderConfig), "provider", serviceConfig.ProviderName, "namespace", serviceConfig.Pod.Namespace)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(
		func() error {
			provider, err := kfp.NewKfpProvider(*kfpProviderConfig)
			if err != nil {
				return fmt.Errorf("failed to create provider: %w", err)
			}
			return server.Start(ctx, *serviceConfig, provider)
		},
	)

	go func() {
		<-ctx.Done()
		logger.Info("context canceled; shutting down...")
	}()

	if err := g.Wait(); err != nil {
		logger.Error(err, "kfp provider crashed")
		os.Exit(1)
	}

	logger.Info("kfp provider terminated gracefully")
}
