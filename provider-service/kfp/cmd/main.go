package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-logr/logr"
	"github.com/go-resty/resty/v2"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	baseConfig "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams/sinks"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams/sources"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	kfpConfig "github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/event"
	kfp "github.com/sky-uk/kfp-operator/provider-service/kfp/internal/provider"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

func main() {
	logger, err := common.NewLogger(zapcore.InfoLevel)
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

	serviceConfig, err := baseConfig.LoadConfig(baseConfig.Config{
		Server: baseConfig.Server{
			Host: "0.0.0.0",
			Port: 8080,
		},
	})
	if err != nil {
		logger.Error(err, "failed to load config")
		panic(err)
	}
	logger.Info(fmt.Sprintf("loaded base config: %+v", serviceConfig))

	kfpProviderConfig, err := baseConfig.LoadConfig(kfpConfig.KfpProviderConfig{Name: serviceConfig.ProviderName, PipelineRootStorage: serviceConfig.PipelineRootStorage})
	if err != nil {
		logger.Error(err, "failed to load provider config", "provider", serviceConfig.ProviderName, "namespace", serviceConfig.Pod.Namespace)
		panic(err)
	}
	logger.Info(fmt.Sprintf("loaded provider config: %+v", kfpProviderConfig), "provider", serviceConfig.ProviderName, "namespace", serviceConfig.Pod.Namespace)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(
		func() error {
			provider, err := kfp.NewKfpProvider(kfpProviderConfig)
			if err != nil {
				return fmt.Errorf("failed to create provider: %w", err)
			}
			return server.Start(ctx, *serviceConfig, provider)
		},
	)

	k8sClient, err := NewK8sClient()
	if err != nil {
		panic(err)
	}

	g.Go(
		func() error {
			runEventing(ctx, *k8sClient, serviceConfig, kfpProviderConfig)
			return nil
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

func runEventing(ctx context.Context, k8sClient K8sClient, baseConfig *baseConfig.Config, providerConfig *kfpConfig.KfpProviderConfig) {
	kfpApi, err := client.CreateKfpApi(ctx, *providerConfig)
	if err != nil {
		panic(err)
	}

	kfpMetadataStore, err := client.CreateMetadataStore(ctx, *providerConfig)
	if err != nil {
		panic(err)
	}

	source, err := sources.NewWorkflowSource(ctx, providerConfig.Parameters.KfpNamespace, k8sClient)
	if err != nil {
		panic(err)
	}

	flow, err := event.NewEventFlow(ctx, *providerConfig, kfpApi, kfpMetadataStore)
	if err != nil {
		panic(err)
	}

	sink, err := sinks.NewObservedWebhookSink(ctx, resty.New(), baseConfig.OperatorWebhook, make(chan StreamMessage[*common.RunCompletionEventData]))
	if err != nil {
		panic(fmt.Errorf("failed to create webhook sink: %w", err))
	}
	errorSink := sinks.NewErrorSink(ctx, make(chan error))

	connectedFlow := flow.From(source)
	connectedFlow.To(sink)
	connectedFlow.Error(errorSink)
}
