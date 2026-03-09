package sources

import (
	"context"
	"encoding/json"
	"fmt"

	pubsub "cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/go-logr/logr"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
)

type PubSubSource struct {
	Logger logr.Logger
	out    chan StreamMessage[string]
	errOut chan error
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

func NewPubSubSource(
	ctx context.Context,
	project string,
	subscription string,
	client *pubsub.Client,
) (*PubSubSource, error) {
	logger := logr.FromContextOrDiscard(ctx)

	fullyQualifiedSub := fmt.Sprintf("projects/%s/subscriptions/%s", project, subscription)

	if _, err := client.SubscriptionAdminClient.GetSubscription(
		ctx,
		&pubsubpb.GetSubscriptionRequest{Subscription: fullyQualifiedSub},
	); err != nil {
		return nil, fmt.Errorf("something went wrong while trying to fetch subscription %s, %s", subscription, err)
	}

	pubSubSource := &PubSubSource{
		Logger: logger,
		out:    make(chan StreamMessage[string]),
		errOut: make(chan error),
	}

	go func() {
		if err := pubSubSource.subscribe(ctx, client.Subscriber(subscription)); err != nil {
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

func (pss *PubSubSource) subscribe(ctx context.Context, subscriber *pubsub.Subscriber) error {
	pss.Logger.Info("subscribing to pubsub...")

	err := subscriber.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		pss.Logger.Info("message received from Pub/Sub", "id", msg.ID)

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
				OnSuccessHandler:              func() { msg.Ack() },
				OnRecoverableFailureHandler:   func() { msg.Nack() },
				OnUnrecoverableFailureHandler: func() { msg.Ack() },
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
