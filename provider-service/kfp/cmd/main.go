package main

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-resty/resty/v2"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	configLoader "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams/sinks"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams/sources"
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
		panic(err)
	}

	k8sClient, err := NewK8sClient()
	if err != nil {
		panic(err)
	}

	providerConfig, err := kfp.LoadProviderConfig(ctx, *k8sClient, config.ProviderName, config.Pod.Namespace)
	if err != nil {
		logger.Error(err, "failed to load provider config", "provider", config.ProviderName, "namespace", config.Pod.Namespace)
		panic(err)
	}

	source, err := sources.NewWorkflowSource(ctx, config.Pod.Namespace, *k8sClient)
	if err != nil {
		panic(err)
	}

	kfpApi, err := kfp.CreateKfpApi(ctx, *providerConfig)
	if err != nil {
		panic(err)
	}

	kfpMetadataStore, err := kfp.CreateMetadataStore(ctx, *providerConfig)
	if err != nil {
		panic(err)
	}

	flow, err := kfp.NewEventFlow(ctx, *providerConfig, kfpApi, kfpMetadataStore)
	if err != nil {
		panic(err)
	}

	sink := sinks.NewWebhookSink(ctx, resty.New(), config.OperatorWebhook, make(chan StreamMessage[*common.RunCompletionEventData]))
	errorSink := sinks.NewErrorSink(ctx, make(chan error))

	connectedFlow := flow.From(source)
	connectedFlow.To(sink)
	connectedFlow.Error(errorSink)
}
