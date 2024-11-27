package internal

import (
	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"context"
	"github.com/go-resty/resty/v2"
	"github.com/reugn/go-streams/flow"
	"github.com/sky-uk/kfp-operator/argo/common"
	k8sApi "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/publisher"
	"google.golang.org/api/option"
	"os"
)

func Start(ctx context.Context, config config.Config) {
	logger := common.LoggerFromContext(ctx)

	k8sClient, err := k8sApi.NewK8sClient()
	if err != nil {
		logger.Error(err, "failed to initialise K8s Client")
		os.Exit(1)
	}

	providerConfig := &VAIProviderConfig{
		Name: config.ProviderName,
	}

	if err = k8sApi.LoadProvider(
		ctx,
		k8sClient.Client,
		config.ProviderName,
		config.Pod.Namespace,
		providerConfig,
	); err != nil {
		logger.Error(err, "failed to load provider", "name", config.ProviderName, "namespace", config.Pod.Namespace)
		os.Exit(1)
	}

	pipelineJobClient, err := aiplatform.NewPipelineClient(ctx, option.WithEndpoint(providerConfig.vaiEndpoint()))
	if err != nil {
		logger.Error(err, "failed to create VAI pipeline client", "endpoint", providerConfig.vaiEndpoint())
		os.Exit(1)
	}
	source, err := NewVaiEventSource(ctx, config.ProviderName, config.Pod.Namespace, pipelineJobClient)
	if err != nil {
		logger.Error(err, "Failed to create VAI event data source")
		os.Exit(1)
	}
	sink := publisher.NewHttpWebhookSink(ctx, config.OperatorWebhook, resty.New(), make(chan any))
	source.Via(flow.NewPassThrough()).To(sink)
}
