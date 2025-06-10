package sinks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type WebhookSink struct {
	client          *resty.Client
	operatorWebhook string
	in              chan StreamMessage[*common.RunCompletionEventData]
}

func NewWebhookSink(ctx context.Context, client *resty.Client, operatorWebhook string, inChan chan StreamMessage[*common.RunCompletionEventData]) *WebhookSink {
	webhookSink := &WebhookSink{client: client, operatorWebhook: operatorWebhook, in: inChan}

	if err := initWebhookMetrics(); err != nil {
		return nil
	}

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
				webhookCounter.Add(ctx, 1, metric.WithAttributes(attribute.String(sendEventsMetricResultKey, RecoverableFailure.String())))
				message.OnRecoverableFailure()
			} else {
				switch response.StatusCode() {
				case http.StatusOK:
					logger.Info("successfully sent event", "event", fmt.Sprintf("%+v", message.Message))
					webhookCounter.Add(ctx, 1, metric.WithAttributes(attribute.String(sendEventsMetricResultKey, Success.String())))
					message.OnSuccess()
				case http.StatusGone:
					logger.Info("resource tied to event is gone", "event", fmt.Sprintf("%+v", message.Message))
					webhookCounter.Add(ctx, 1, metric.WithAttributes(attribute.String(sendEventsMetricResultKey, UnrecoverableFailure.String())))
					message.OnUnrecoverableFailureHandler()
				default:
					logger.Error(fmt.Errorf("unexpected status code %d", response.StatusCode()), "unexpected response from webhook", "event", fmt.Sprintf("%+v", message.Message))
					webhookCounter.Add(ctx, 1, metric.WithAttributes(attribute.String(sendEventsMetricResultKey, RecoverableFailure.String())))
					message.OnRecoverableFailureHandler()
				}
			}
		} else {
			logger.Info("discarding empty message")
			webhookCounter.Add(ctx, 1, metric.WithAttributes(attribute.String(sendEventsMetricResultKey, Discarded.String())))
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

var webhookCounter metric.Int64Counter

func initWebhookMetrics() error {
	meter := otel.Meter("webhook-sink")

	var err error
	webhookCounter, err = meter.Int64Counter("provider_webhook_send_events_count")

	return err
}
