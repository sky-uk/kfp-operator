package sources

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	"os"
)

type PubSubSource struct {
	RunsSubscription *pubsub.Subscription
	Logger           logr.Logger
	out              chan StreamMessage[string]
}

type Resource struct {
	Labels map[string]string `json:"labels"`
}

type LogEntry struct {
	Resource Resource `json:"resource"`
}

func NewPubSubSource(ctx context.Context, project string, subscription string) (*PubSubSource, error) {
	logger := common.LoggerFromContext(ctx)

	pubSubClient, err := pubsub.NewClient(ctx, project)
	if err != nil {
		logger.Error(err, "failed to create pubsub client", "project", project)
		return nil, err
	}

	runsSubscription := pubSubClient.Subscription(subscription)

	pubSubSource := &PubSubSource{
		Logger: logger,
		out:    make(chan StreamMessage[string]),
	}

	go func() {
		if err := pubSubSource.subscribe(ctx, runsSubscription); err != nil {
			logger.Error(err, "Failed to subscribe", "subscription", subscription)
			os.Exit(1)
		}
	}()

	return pubSubSource, nil
}

func (pss *PubSubSource) Out() <-chan StreamMessage[string] {
	return pss.out
}

func (pss *PubSubSource) subscribe(ctx context.Context, runsSubscription *pubsub.Subscription) error {
	pss.Logger.Info("subscribing to pubsub...")

	err := runsSubscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		pss.Logger.Info(fmt.Sprintf("message received from Pub/Sub with ID: %s", msg.ID))

		pipelineJobId, err := pss.extractPipelineJobId(msg)
		if err != nil {
			pss.Logger.Error(err, "failed to extract pipeline_job_id from message")
			msg.Nack()
		}

		select {
		case pss.out <- StreamMessage[string]{
			Message: pipelineJobId,
			OnCompleteHandlers: OnCompleteHandlers{
				OnSuccessHandler: func() { msg.Ack() },
				OnFailureHandler: func() { msg.Nack() },
			},
		}:
		case <-ctx.Done():
			pss.Logger.Info("stopped reading from pubsub")
			return
		}
	})

	if err != nil {
		pss.Logger.Error(err, "failed to read from pubsub")
		return err
	}

	return nil
}

func (pss *PubSubSource) extractPipelineJobId(msg *pubsub.Message) (string, error) {
	logEntry := LogEntry{}
	err := json.Unmarshal(msg.Data, &logEntry)
	if err != nil {
		return "", err
	}

	pipelineJobId, ok := logEntry.Resource.Labels["pipeline_job_id"]
	if !ok {
		err := fmt.Errorf("logEntry did not contain pipeline_job_id %+v", logEntry)
		return "", err
	}
	return pipelineJobId, nil
}
