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
	ctx             context.Context
	client          *resty.Client
	operatorWebhook string
	in              chan StreamMessage[*common.RunCompletionEventData]
}

func NewWebhookSink(ctx context.Context, client *resty.Client, operatorWebhook string, inChan chan StreamMessage[*common.RunCompletionEventData]) *WebhookSink {
	webhookSink := &WebhookSink{ctx: ctx, client: client, operatorWebhook: operatorWebhook, in: inChan}

	go webhookSink.SendEvents()

	return webhookSink
}

func (hws WebhookSink) In() chan<- StreamMessage[*common.RunCompletionEventData] {
	return hws.in
}

func (hws WebhookSink) SendEvents() {
	logger := common.LoggerFromContext(hws.ctx)
	for message := range hws.in {
		if message.Message != nil {
			err, httpCode := hws.send(*message.Message)
			if err != nil {
				logger.Error(err, "failed to send event", "event", fmt.Sprintf("%+v", message.Message))
				switch httpCode {
				case http.StatusGone:
					message.OnUnrecoverableFailure()
				default:
					message.OnRecoverableFailure()
				}
			} else {
				logger.Info("successfully sent event", "event", fmt.Sprintf("%+v", message.Message))
				message.OnSuccess()
			}
		} else {
			logger.Info("discarding empty message")
		}
	}
}

func (hws WebhookSink) send(rced common.RunCompletionEventData) (error, int) {
	rcedBytes, err := json.Marshal(rced)
	if err != nil {
		return err, 0
	}

	response, err := hws.client.R().SetHeader("Content-Type", "application/json").SetBody(rcedBytes).Post(hws.operatorWebhook)
	if err != nil {
		return err, 0
	}

	if response.StatusCode() != http.StatusOK {
		logger := common.LoggerFromContext(hws.ctx)
		logger.Info("error returned from Webhook", "status", response.Status(), "body", response.Body(), "event", rced)
		return fmt.Errorf("webhook error response received with http status code: [%s]", response.Status()), response.StatusCode()
	}

	return nil, http.StatusOK
}
