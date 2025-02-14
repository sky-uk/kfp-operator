package sources

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/go-logr/logr"
	"github.com/sky-uk/kfp-operator/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
)

type PubSubSource struct {
	RunsSubscription *pubsub.Subscription
	Logger           logr.Logger
	out              chan StreamMessage[string]
	errOut           chan error
}

type Resource struct {
	Labels map[string]string `json:"labels"`
}

type LogEntry struct {
	Resource Resource `json:"resource"`
}

const PipelineJobLabel = "pipeline_job_id"

func (ps *PubSubSource) ErrorOut() <-chan error {
	return ps.errOut
}

func NewPubSubSource(ctx context.Context, project string, subscription string) (*PubSubSource, error) {
	logger := common.LoggerFromContext(ctx)

	pubSubClient, err := pubsub.NewClient(ctx, project)
	if err != nil {
		logger.Error(err, "failed to create pubsub client", "project", project)
		return nil, err
	}

	runsSubscription := pubSubClient.Subscription(subscription)
	exists, err := runsSubscription.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while trying to fetch subscription %s, %s", subscription, err)
	} else if !exists {
		return nil, fmt.Errorf("subscription %s does not exist on topic", subscription)
	}

	pubSubSource := &PubSubSource{
		Logger: logger,
		out:    make(chan StreamMessage[string]),
		errOut: make(chan error),
	}

	go func() {
		if err := pubSubSource.subscribe(ctx, runsSubscription); err != nil {
			logger.Error(err, "failed to subscribe", "subscription", subscription)
			pubSubSource.errOut <- err
			return
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
			return
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

	pipelineJobId, ok := logEntry.Resource.Labels[PipelineJobLabel]
	if !ok {
		err := fmt.Errorf("logEntry did not contain %s in %+v", PipelineJobLabel, logEntry)
		return "", err
	}
	return pipelineJobId, nil
}
