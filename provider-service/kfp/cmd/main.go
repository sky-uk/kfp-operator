package main

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-resty/resty/v2"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	baseConfig "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams/sinks"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams/sources"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/event"
	kfp "github.com/sky-uk/kfp-operator/provider-service/kfp/internal/provider"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger, err := common.NewLogger(zapcore.InfoLevel)
	if err != nil {
		panic(err)
	}

	ctx := logr.NewContext(context.Background(), logger)

	baseConfig, err := baseConfig.LoadConfig(ctx)
	if err != nil {
		panic(err)
	}
	logger.Info(fmt.Sprintf("loaded base config: %+v", baseConfig))

	kfpProviderConfig, err := config.LoadKfpProviderConfig(baseConfig.ProviderName)
	if err != nil {
		logger.Error(err, "failed to load provider config", "provider", baseConfig.ProviderName, "namespace", baseConfig.Pod.Namespace)
		panic(err)
	}
	logger.Info(fmt.Sprintf("loaded provider config: %+v", kfpProviderConfig), "provider", baseConfig.ProviderName, "namespace", baseConfig.Pod.Namespace)

	go runServer(ctx, kfpProviderConfig, baseConfig)

	k8sClient, err := NewK8sClient()
	if err != nil {
		panic(err)
	}

	go runEventing(ctx, *k8sClient, baseConfig, kfpProviderConfig)

	<-ctx.Done()
}

func runServer(ctx context.Context, kfpConfig *config.KfpProviderConfig, baseConfig *baseConfig.Config) {
	provider, err := kfp.NewKfpProvider(ctx, kfpConfig)
	if err != nil {
		panic(err)
	}

	if err = server.Start(ctx, baseConfig.Server, provider); err != nil {
		panic(err)
	}
}

func runEventing(ctx context.Context, k8sClient K8sClient, baseConfig *baseConfig.Config, providerConfig *config.KfpProviderConfig) {
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

	sink := sinks.NewWebhookSink(ctx, resty.New(), baseConfig.OperatorWebhook, make(chan StreamMessage[*common.RunCompletionEventData]))
	errorSink := sinks.NewErrorSink(ctx, make(chan error))

	connectedFlow := flow.From(source)
	connectedFlow.To(sink)
	connectedFlow.Error(errorSink)
}
