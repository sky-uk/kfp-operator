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
	SendEvents(ctx context.Context, handler MessageHandler)
}

type MessageHandler interface {
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

	go webhookSink.SendEvents(ctx, webhookSink)

	return webhookSink
}

func NewObservedWebhookSink(ctx context.Context, client *resty.Client, operatorWebhook string, inChan chan StreamMessage[*common.RunCompletionEventData]) (*ObservedWebhookSink, error) {
	meter := otel.Meter("webhook_sink")
	sendEventsCounter, err := meter.Int64Counter(
		"provider_webhook_send_events_count",
		metric.WithDescription("Total number of provider webhook sink SendEvents calls"),
	)
	if err != nil {
		common.LoggerFromContext(ctx).Error(err, "failed to create webhook counter")
		return nil, err
	}

	delegateWebSink := WebhookSink{client: client, operatorWebhook: operatorWebhook, in: inChan}
	observed := ObservedWebhookSink{
		delegate:          delegateWebSink,
		meter:             meter,
		sendEventsCounter: sendEventsCounter,
	}

	go delegateWebSink.SendEvents(ctx, observed)
	return &observed, nil
}

func (ws WebhookSink) OnEmpty(ctx context.Context) {}
func (ws WebhookSink) OnError(ctx context.Context, message StreamMessage[*common.RunCompletionEventData]) {
	message.OnRecoverableFailure()
}
func (ws WebhookSink) OnRecoverableFailure(ctx context.Context, message StreamMessage[*common.RunCompletionEventData]) {
	message.OnRecoverableFailure()
}

func (ws WebhookSink) OnUnrecoverableFailure(ctx context.Context, message StreamMessage[*common.RunCompletionEventData]) {
	message.OnUnrecoverableFailure()
}

func (ws WebhookSink) OnSuccess(ctx context.Context, message StreamMessage[*common.RunCompletionEventData]) {
	message.OnSuccess()
}

func (ws WebhookSink) In() chan<- StreamMessage[*common.RunCompletionEventData] {
	return ws.in
}

func (ws WebhookSink) SendEvents(ctx context.Context, handler MessageHandler) {
	for message := range ws.in {
		if message.Message != nil {
			err, response := ws.send(*message.Message)
			if err != nil || response == nil {
				handler.OnError(ctx, message)
			} else {
				switch response.StatusCode() {
				case http.StatusOK:
					handler.OnSuccess(ctx, message)
				case http.StatusGone:
					handler.OnUnrecoverableFailure(ctx, message)
				default:
					handler.OnRecoverableFailure(ctx, message)
				}
			}
		} else {
			handler.OnEmpty(ctx)
		}
	}
}

func (ws WebhookSink) send(rced common.RunCompletionEventData) (error, *resty.Response) {
	rcedBytes, err := json.Marshal(rced)
	if err != nil {
		return err, nil
	}

	response, err := ws.client.R().SetHeader("Content-Type", "application/json").SetBody(rcedBytes).Post(ws.operatorWebhook)
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

func (ows ObservedWebhookSink) SendEvents(ctx context.Context, _ MessageHandler) {
	ows.delegate.SendEvents(ctx, ows)
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
	jsonMessage, _ := json.Marshal(message.Message)
	logger.Error(fmt.Errorf("webhook sink received error message"), "", "message", string(jsonMessage))
	ows.sendEventsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String(sendEventsMetricResultKey, RecoverableFailure.String())))
	ows.delegate.OnError(ctx, message)
}

func (ows ObservedWebhookSink) OnRecoverableFailure(ctx context.Context, message StreamMessage[*common.RunCompletionEventData]) {
	logger := common.LoggerFromContext(ctx)
	jsonMessage, _ := json.Marshal(message.Message)
	logger.Error(fmt.Errorf("webhook sink received recoverable failure"), "", "message", string(jsonMessage))
	ows.sendEventsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String(sendEventsMetricResultKey, RecoverableFailure.String())))
	ows.delegate.OnRecoverableFailure(ctx, message)
}

func (ows ObservedWebhookSink) OnUnrecoverableFailure(ctx context.Context, message StreamMessage[*common.RunCompletionEventData]) {
	logger := common.LoggerFromContext(ctx)
	jsonMessage, _ := json.Marshal(message.Message)
	logger.Error(fmt.Errorf("webhook sink received unrecoverable failure"), "", "message", string(jsonMessage))
	ows.sendEventsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String(sendEventsMetricResultKey, UnrecoverableFailure.String())))
	ows.delegate.OnUnrecoverableFailure(ctx, message)
}

func (ows ObservedWebhookSink) OnSuccess(ctx context.Context, message StreamMessage[*common.RunCompletionEventData]) {
	logger := common.LoggerFromContext(ctx)
	logger.Info("webhook sink received successful", "runId", message.Message.RunId)
	ows.sendEventsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String(sendEventsMetricResultKey, Success.String())))
	ows.delegate.OnSuccess(ctx, message)
}
