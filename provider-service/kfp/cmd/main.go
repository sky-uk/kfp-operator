package main

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-resty/resty/v2"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	baseConfigLoader "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams/sinks"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams/sources"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/event"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger, err := common.NewLogger(zapcore.InfoLevel)
	if err != nil {
		panic(err)
	}

	ctx := logr.NewContext(context.Background(), logger)

	baseConfig, err := baseConfigLoader.LoadConfig(ctx)
	if err != nil {
		panic(err)
	}

	k8sClient, err := NewK8sClient()
	if err != nil {
		panic(err)
	}

	providerConfig, err := config.LoadProviderConfig(ctx, *k8sClient, baseConfig.ProviderName, baseConfig.Pod.Namespace)
	if err != nil {
		logger.Error(err, "failed to load provider config", "provider", baseConfig.ProviderName, "namespace", baseConfig.Pod.Namespace)
		panic(err)
	}

	source, err := sources.NewWorkflowSource(ctx, providerConfig.Parameters.KfpNamespace, *k8sClient)
	if err != nil {
		panic(err)
	}

	kfpApi, err := client.CreateKfpApi(ctx, *providerConfig)
	if err != nil {
		panic(err)
	}

	kfpMetadataStore, err := client.CreateMetadataStore(ctx, *providerConfig)
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
