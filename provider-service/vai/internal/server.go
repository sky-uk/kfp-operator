package internal

import (
	"context"
	"github.com/go-resty/resty/v2"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/sinks"
	"os"
)

func Start(ctx context.Context, config config.Config) {
	logger := common.LoggerFromContext(ctx)
	source, err := NewVaiEventSource(ctx, config.ProviderName, config.Pod.Namespace)
	if err != nil {
		logger.Error(err, "Failed to create VAI event data source")
		os.Exit(1)
	}
	sink := sinks.NewWebhookSink(ctx, config.OperatorWebhook, resty.New(), make(chan any))
	source.Via(source.RunCompletionEventConversionFlow).To(sink)
}
