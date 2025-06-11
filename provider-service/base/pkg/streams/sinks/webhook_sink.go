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

type MessageSink interface {
	In() chan<- StreamMessage[*common.RunCompletionEventData]
	SendEvents(ctx context.Context)
	OnEmpty(ctx context.Context)
	OnError(ctx context.Context, message StreamMessage[*common.RunCompletionEventData])
	OnRecoverableFailure(ctx context.Context, message StreamMessage[*common.RunCompletionEventData])
	OnUnrecoverableFailure(ctx context.Context, message StreamMessage[*common.RunCompletionEventData])
	OnSuccess(ctx context.Context, message StreamMessage[*common.RunCompletionEventData])
}

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

func (hws WebhookSink) WithMetrics(ctx context.Context) (*ObservedWebhookSink, error) {
	meter := otel.Meter("webhook_sink")
	sendEventsCounter, err := meter.Int64Counter(
		"provider_webhook_send_events_count",
		metric.WithDescription("Total number of provider webhook sink SendEvents calls"),
	)
	if err != nil {
		common.LoggerFromContext(ctx).Error(err, "failed to create webhook counter")
		return nil, err
	}

	return &ObservedWebhookSink{
		delegate:          hws,
		meter:             meter,
		sendEventsCounter: sendEventsCounter,
	}, nil
}

func (hws WebhookSink) OnEmpty(ctx context.Context) {}
func (hws WebhookSink) OnError(ctx context.Context, message StreamMessage[*common.RunCompletionEventData]) {
	message.OnRecoverableFailure()
}
func (hws WebhookSink) OnRecoverableFailure(ctx context.Context, message StreamMessage[*common.RunCompletionEventData]) {
	message.OnRecoverableFailure()
}

func (hws WebhookSink) OnUnrecoverableFailure(ctx context.Context, message StreamMessage[*common.RunCompletionEventData]) {
	message.OnUnrecoverableFailure()
}

func (hws WebhookSink) OnSuccess(ctx context.Context, message StreamMessage[*common.RunCompletionEventData]) {
	message.OnSuccess()
}

func (hws WebhookSink) In() chan<- StreamMessage[*common.RunCompletionEventData] {
	return hws.in
}

func (hws WebhookSink) SendEvents(ctx context.Context) {
	for message := range hws.in {
		if message.Message != nil {
			err, response := hws.send(*message.Message)
			if err != nil || response == nil {
				hws.OnError(ctx, message)
			} else {
				switch response.StatusCode() {
				case http.StatusOK:
					hws.OnSuccess(ctx, message)
				case http.StatusGone:
					hws.OnUnrecoverableFailure(ctx, message)
				default:
					hws.OnRecoverableFailure(ctx, message)
				}
			}
		} else {
			hws.OnEmpty(ctx)
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

type ObservedWebhookSink struct {
	delegate          WebhookSink
	meter             metric.Meter
	sendEventsCounter metric.Int64Counter
}

func (ows ObservedWebhookSink) SendEvents(ctx context.Context) {
	ows.delegate.SendEvents(ctx)
}

func (ows ObservedWebhookSink) In() chan<- StreamMessage[*common.RunCompletionEventData] {
	return ows.delegate.in
}

func (ows ObservedWebhookSink) OnEmpty(ctx context.Context) {
	logger := common.LoggerFromContext(ctx)
	logger.Info("webhook sink received empty message")
	ows.sendEventsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String(sendEventsMetricResultKey, Discarded.String())))
	ows.delegate.OnEmpty(ctx)
}

func (ows ObservedWebhookSink) OnError(ctx context.Context, message StreamMessage[*common.RunCompletionEventData]) {
	logger := common.LoggerFromContext(ctx)
	logger.Error(fmt.Errorf("webhook sink received error message"), "error message", fmt.Sprintf("%+v", message))
	ows.sendEventsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String(sendEventsMetricResultKey, RecoverableFailure.String())))
	ows.delegate.OnError(ctx, message)
}

func (ows ObservedWebhookSink) OnRecoverableFailure(ctx context.Context, message StreamMessage[*common.RunCompletionEventData]) {
	logger := common.LoggerFromContext(ctx)
	logger.Error(fmt.Errorf("webhook sink received recoverable failure"), "message", fmt.Sprintf("%+v", message))
	ows.sendEventsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String(sendEventsMetricResultKey, RecoverableFailure.String())))
	ows.delegate.OnRecoverableFailure(ctx, message)
}

func (ows ObservedWebhookSink) OnUnrecoverableFailure(ctx context.Context, message StreamMessage[*common.RunCompletionEventData]) {
	logger := common.LoggerFromContext(ctx)
	logger.Error(fmt.Errorf("webhook sink received unrecoverable failure"), "message", fmt.Sprintf("%+v", message))
	ows.sendEventsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String(sendEventsMetricResultKey, UnrecoverableFailure.String())))
	ows.delegate.OnUnrecoverableFailure(ctx, message)
}

func (ows ObservedWebhookSink) OnSuccess(ctx context.Context, message StreamMessage[*common.RunCompletionEventData]) {
	logger := common.LoggerFromContext(ctx)
	logger.Info("webhook sink received success message", "message", fmt.Sprintf("%+v", message))
	ows.sendEventsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String(sendEventsMetricResultKey, Success.String())))
	ows.delegate.OnSuccess(ctx, message)
}
