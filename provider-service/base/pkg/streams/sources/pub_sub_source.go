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
	Labels   map[string]string `json:"labels"`
	Resource Resource          `json:"resource"`
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

func (es *PubSubSource) Out() <-chan StreamMessage[string] {
	return es.out
}

func (es *PubSubSource) subscribe(ctx context.Context, runsSubscription *pubsub.Subscription) error {
	es.Logger.Info("subscribing to pubsub...")

	err := runsSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		es.Logger.Info(fmt.Sprintf("message received from Pub/Sub with ID: %s", m.ID))
		logEntry := LogEntry{}
		err := json.Unmarshal(m.Data, &logEntry)
		if err != nil {
			es.Logger.Error(err, "failed to unmarshal Pub/Sub message")
			m.Nack()
			return
		}
		es.Logger.Info(fmt.Sprintf("%+v", logEntry))

		pipelineJobId, ok := logEntry.Resource.Labels["pipeline_job_id"]
		if !ok {
			es.Logger.Error(err, fmt.Sprintf("logEntry did not contain pipeline_job_id %+v", logEntry))
			m.Nack()
			return
		}

		select {
		case es.out <- StreamMessage[string]{
			Message: pipelineJobId,
			OnCompleteHandlers: OnCompleteHandlers{
				OnSuccessHandler: func() { m.Ack() },
				OnFailureHandler: func() { m.Nack() },
			},
		}:
		case <-ctx.Done():
			es.Logger.Info("stopped reading from pubsub")
			return
		}
	})

	if err != nil {
		es.Logger.Error(err, "failed to read from pubsub")
		return err
	}

	return nil
}
