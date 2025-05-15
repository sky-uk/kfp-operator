package main

import (
	"context"
	"fmt"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/client"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"github.com/go-logr/logr"
	"github.com/go-resty/resty/v2"
	"github.com/sky-uk/kfp-operator/argo/common"
	baseConfig "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams/sinks"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams/sources"
	vaiConfig "github.com/sky-uk/kfp-operator/provider-service/vai/internal/config"
	vai "github.com/sky-uk/kfp-operator/provider-service/vai/internal/provider"
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

	serviceConfig, err := baseConfig.LoadConfig(baseConfig.Config{
		Server: baseConfig.Server{
			Host: "0.0.0.0",
			Port: 8080,
		},
	})
	if err != nil {
		panic(err)
	}
	logger.Info(fmt.Sprintf("loaded base config: %+v", serviceConfig))

	vaiProviderConfig, err := baseConfig.LoadConfig(vaiConfig.VAIProviderConfig{Name: serviceConfig.ProviderName, PipelineRootStorage: serviceConfig.PipelineRootStorage})
	if err != nil {
		logger.Error(err, "failed to load provider config", "provider", serviceConfig.ProviderName, "namespace", serviceConfig.Pod.Namespace)
		panic(err)
	}
	logger.Info(fmt.Sprintf("loaded provider config: %+v", vaiProviderConfig), "provider", serviceConfig.ProviderName, "namespace", serviceConfig.Pod.Namespace)

	pipelineJobClient, err := aiplatform.NewPipelineClient(ctx, option.WithEndpoint(vaiProviderConfig.VaiEndpoint()))
	if err != nil {
		logger.Error(err, "failed to create VAI pipeline client", "endpoint", vaiProviderConfig.VaiEndpoint())
		panic(err)
	}

	scheduleClient, err := aiplatform.NewScheduleClient(
		ctx,
		option.WithEndpoint(vaiProviderConfig.VaiEndpoint()),
	)
	if err != nil {
		logger.Error(err, "failed to create VAI schedule client", "endpoint", vaiProviderConfig.VaiEndpoint())
		panic(err)
	}

	provider, err := vai.NewVAIProvider(ctx, vaiProviderConfig, serviceConfig.Pod.Namespace, pipelineJobClient, scheduleClient)
	if err != nil {
		logger.Error(err, "failed to create VAI Provider", "endpoint", vaiProviderConfig.VaiEndpoint())
		panic(err)
	}

	eventSource, err := sources.NewPubSubSource(ctx, vaiProviderConfig.Parameters.VaiProject, vaiProviderConfig.Parameters.EventsourcePipelineEventsSubscription)
	if err != nil {
		logger.Error(err, "failed to create VAI event data source")
		panic(err)
	}

	go func() {
		if err = server.Start(ctx, *serviceConfig, provider, []server.HealthCheck{
			provider,
			eventSource,
		}); err != nil {
			panic(err)
		}
	}()

	runEventing(ctx, logger, serviceConfig, vaiProviderConfig, eventSource, pipelineJobClient)

	<-ctx.Done()
	logger.Info("vai event flow is terminating")
}

func runEventing(ctx context.Context, logger logr.Logger, baseConfig *baseConfig.Config, providerConfig *vaiConfig.VAIProviderConfig, source *sources.PubSubSource, pipelineJobClient client.PipelineJobClient) {
	go handleErrorInSourceOperations(source)

	flow := runcompletion.NewEventFlow(providerConfig, pipelineJobClient)

	go func() {
		flow.Start(ctx)
	}()

	sink := sinks.NewWebhookSink(ctx, resty.New(), baseConfig.OperatorWebhook, make(chan StreamMessage[*common.RunCompletionEventData]))
	errorSink := sinks.NewErrorSink(ctx, make(chan error))

	logger.Info("starting vai event flow")
	flow.From(source).To(sink)
	flow.Error(errorSink)
}

func handleErrorInSourceOperations(source *sources.PubSubSource) {
	for err := range source.ErrorOut() {
		panic(err)
	}
}
