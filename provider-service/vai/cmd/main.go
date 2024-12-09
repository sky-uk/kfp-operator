package main

import (
	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"context"
	"github.com/go-logr/logr"
	"github.com/go-resty/resty/v2"
	"github.com/sky-uk/kfp-operator/argo/common"
	configLoader "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams/sinks"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams/sources"
	vai "github.com/sky-uk/kfp-operator/provider-service/vai/internal"
	"go.uber.org/zap/zapcore"

	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"

	"google.golang.org/api/option"
	"os"
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
		os.Exit(1)
	}

	k8sClient, err := NewK8sClient()
	if err != nil {
		logger.Error(err, "failed to initialise K8s Client")
		os.Exit(1)
	}

	vaiConfig := &vai.VAIProviderConfig{
		Name: config.ProviderName,
	}

	if err := LoadProvider(ctx, k8sClient.Client, config.ProviderName, config.Pod.Namespace, vaiConfig); err != nil {
		logger.Error(err, "failed to load provider", "name", config.ProviderName, "namespace", config.Pod.Namespace)
		os.Exit(1)
	}

	pipelineJobClient, err := aiplatform.NewPipelineClient(ctx, option.WithEndpoint(vaiConfig.VaiEndpoint()))
	if err != nil {
		logger.Error(err, "failed to create VAI pipeline client", "endpoint", vaiConfig.VaiEndpoint())
		os.Exit(1)
	}

	source, err := sources.NewPubSubSource(ctx, vaiConfig.Parameters.VaiProject, vaiConfig.Parameters.EventsourcePipelineEventsSubscription)
	if err != nil {
		logger.Error(err, "Failed to create VAI event data source")
		os.Exit(1)
	}
	go handleErrorInSourceOperations(source)

	flow := vai.NewVaiEventFlow(ctx, vaiConfig, pipelineJobClient)

	sink := sinks.NewWebhookSink(ctx, resty.New(), config.OperatorWebhook, make(chan StreamMessage[*common.RunCompletionEventData]))

	flow.From(source).To(sink)
}

func handleErrorInSourceOperations(source *sources.PubSubSource) {
	for range source.ErrorOut() {
		os.Exit(1)
	}
}
