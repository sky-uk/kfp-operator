package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
)

type HttpWebhookSink struct {
	context         context.Context
	client          *resty.Client
	operatorWebhook string
	in              chan StreamMessage[*common.RunCompletionEventData]
}

func NewHttpWebhookSink(ctx context.Context, operatorWebhook string, client *resty.Client, inChan chan StreamMessage[*common.RunCompletionEventData]) *HttpWebhookSink {
	httpWebhook := &HttpWebhookSink{context: ctx, client: client, operatorWebhook: operatorWebhook, in: inChan}

	go httpWebhook.SendEvents()

	return httpWebhook
}

func (hws HttpWebhookSink) In() chan<- StreamMessage[*common.RunCompletionEventData] {
	return hws.in
}

func (hws HttpWebhookSink) SendEvents() {
	logger := common.LoggerFromContext(hws.context)
	for message := range hws.in {
		var err error
		if message.Message != nil {
			err = hws.send(*message.Message)
			if err != nil {
				logger.Error(err, "Failed to send event", "event", fmt.Sprintf("%+v", message))
				message.OnFailure()
			} else {
				logger.Info("Successfully sent event", "event", fmt.Sprintf("%+v", message))
				message.OnSuccess()
			}
		} else {
			logger.Info("Discarding empty message")
		}
	}
}

func (hws HttpWebhookSink) send(rced common.RunCompletionEventData) error {
	rcedBytes, err := json.Marshal(rced)
	if err != nil {
		return err
	}

	response, err := hws.client.R().SetHeader("Content-Type", "application/json").SetBody(rcedBytes).Post(hws.operatorWebhook)
	if err != nil {
		return err
	}

	if response.StatusCode() != 200 {
		logger := common.LoggerFromContext(hws.context)
		logger.Info("Error returned from Webhook", "status", response.Status(), "body", response.Body(), "event", rced)
		return fmt.Errorf("KFP Operator error response received with http status code: [%s]", response.Status())
	}

	return nil
}
