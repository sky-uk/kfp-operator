package main

import (
	"context"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"github.com/go-logr/logr"
	"github.com/go-resty/resty/v2"
	"github.com/sky-uk/kfp-operator/argo/common"
	configLoader "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams/sinks"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams/sources"
	vaiConfig "github.com/sky-uk/kfp-operator/provider-service/vai/internal/config"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/runcompletion"
	"go.uber.org/zap/zapcore"

	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"

	"google.golang.org/api/option"
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

	vaiConfig := &vaiConfig.VAIProviderConfig{
		Name: config.ProviderName,
	}

	if err := LoadProvider(ctx, k8sClient.Client, config.ProviderName, config.Pod.Namespace, vaiConfig); err != nil {
		logger.Error(err, "failed to load provider", "name", config.ProviderName, "namespace", config.Pod.Namespace)
		panic(err)
	}

	pipelineJobClient, err := aiplatform.NewPipelineClient(ctx, option.WithEndpoint(vaiConfig.VaiEndpoint()))
	if err != nil {
		logger.Error(err, "failed to create VAI pipeline client", "endpoint", vaiConfig.VaiEndpoint())
		panic(err)
	}

	source, err := sources.NewPubSubSource(ctx, vaiConfig.Parameters.VaiProject, vaiConfig.Parameters.EventsourcePipelineEventsSubscription)
	if err != nil {
		logger.Error(err, "failed to create VAI event data source")
		panic(err)
	}
	go handleErrorInSourceOperations(source)

	flow := runcompletion.NewEventFlow(ctx, vaiConfig, pipelineJobClient)

	go func() {
		flow.Start()
	}()

	sink := sinks.NewWebhookSink(ctx, resty.New(), config.OperatorWebhook, make(chan StreamMessage[*common.RunCompletionEventData]))
	errorSink := sinks.NewErrorSink(ctx, make(chan error))

	logger.Info("starting vai event flow")
	flow.From(source).To(sink)
	flow.Error(errorSink)

	// block till terminated
	<-ctx.Done()
	logger.Info("vai event flow is terminating")
}

func handleErrorInSourceOperations(source *sources.PubSubSource) {
	for err := range source.ErrorOut() {
		panic(err)
	}
}
