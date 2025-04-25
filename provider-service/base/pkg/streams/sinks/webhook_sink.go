package sinks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
)

type WebhookSink struct {
	client          *resty.Client
	operatorWebhook string
	in              chan StreamMessage[*common.RunCompletionEventData]
}

func NewWebhookSink(ctx context.Context, client *resty.Client, operatorWebhook string, inChan chan StreamMessage[*common.RunCompletionEventData]) *WebhookSink {
	webhookSink := &WebhookSink{client: client, operatorWebhook: operatorWebhook, in: inChan}

	go webhookSink.SendEvents(ctx)

	return webhookSink
}

func (hws WebhookSink) In() chan<- StreamMessage[*common.RunCompletionEventData] {
	return hws.in
}

func (hws WebhookSink) SendEvents(ctx context.Context) {
	logger := common.LoggerFromContext(ctx)
	for message := range hws.in {
		if message.Message != nil {
			err, response := hws.send(*message.Message)

			if err != nil || response == nil {
				logger.Error(err, "failed to send event", "event", fmt.Sprintf("%+v", message.Message))
				message.OnRecoverableFailure()
			} else {
				switch response.StatusCode() {
				case http.StatusOK:
					logger.Info("successfully sent event", "event", fmt.Sprintf("%+v", message.Message))
					message.OnSuccess()
				case http.StatusGone:
					logger.Info("resource tied to event is gone", "event", fmt.Sprintf("%+v", message.Message))
					message.OnUnrecoverableFailureHandler()
				default:
					logger.Error(fmt.Errorf("unexpected status code %d", response.StatusCode()), "unexpected response from webhook", "event", fmt.Sprintf("%+v", message.Message))
					message.OnRecoverableFailureHandler()
				}
			}
		} else {
			logger.Info("discarding empty message")
		}
	}
}

func (hws WebhookSink) send(rced common.RunCompletionEventData) (error, *resty.Response) {
	rcedBytes, err := json.Marshal(rced)
	if err != nil {
		return err, nil
	}

	response, err := hws.client.R().SetHeader("Content-Type", "application/json").SetBody(rcedBytes).Post(hws.operatorWebhook)
	if err != nil {
		return err, nil
	}

	return nil, response
}
