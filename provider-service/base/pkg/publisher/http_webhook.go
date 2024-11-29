package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg"
)

type HttpWebhookSink struct {
	context         context.Context
	client          *resty.Client
	operatorWebhook string
	in              chan any
}

func NewHttpWebhookSink(ctx context.Context, operatorWebhook string, client *resty.Client, inChan chan any) *HttpWebhookSink {
	httpWebhook := &HttpWebhookSink{context: ctx, client: client, operatorWebhook: operatorWebhook, in: inChan}

	go httpWebhook.SendEvents()

	return httpWebhook
}

func (hws HttpWebhookSink) In() chan<- any {
	return hws.in
}

func (hws HttpWebhookSink) SendEvents() {
	logger := common.LoggerFromContext(hws.context)
	for data := range hws.in {
		var err error
		switch object := data.(type) {
		case pkg.StreamMessage[*common.RunCompletionEventData]:
			if object.Message != nil {
				err = hws.send(*object.Message)
				if err != nil {
					logger.Error(err, "Failed to send event", "event", fmt.Sprintf("%+v", object))
					object.OnFailure()
				} else {
					logger.Info("Successfully sent event", "event", fmt.Sprintf("%+v", object))
					object.OnSuccess()
				}
			} else {
				logger.Info("Discarding empty message")
			}
		default:
			logger.Info("Unknown object type in stream", "unknown", fmt.Sprintf("%+v", object))
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
