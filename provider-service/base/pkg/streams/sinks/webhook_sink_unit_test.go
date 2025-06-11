//go:build unit

package sinks

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	"go.uber.org/zap/zapcore"
)

func TestSinksUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sinks Unit Suite")
}

var _ = Context("SendEvents", func() {
	var logger, _ = common.NewLogger(zapcore.DebugLevel)
	var ctx = logr.NewContext(context.Background(), logger)
	runCompletionEventData := common.RunCompletionEventData{}

	handlerCall := make(chan any, 1)
	onCompHandlers := OnCompleteHandlers{
		OnSuccessHandler: func() {
			handlerCall <- "success_called"
		},
		OnRecoverableFailureHandler: func() {
			handlerCall <- "recoverable_failure_called"
		},
		OnUnrecoverableFailureHandler: func() {
			handlerCall <- "unrecoverable_failure_called"
		},
	}

	When("receives an http OK response code after sending a message to the webhook server", func() {
		It("should call the message's OnSuccessHandler", func() {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			webhookUrl := "/operator-webhook"
			httpmock.RegisterResponder(http.MethodPost, webhookUrl, httpmock.NewStringResponder(http.StatusOK, ""))

			in := make(chan StreamMessage[*common.RunCompletionEventData])
			webhookSink := &WebhookSink{client: client, operatorWebhook: webhookUrl, in: in}

			go webhookSink.SendEvents(ctx, webhookSink)

			streamMessage := StreamMessage[*common.RunCompletionEventData]{
				Message:            &runCompletionEventData,
				OnCompleteHandlers: onCompHandlers,
			}
			in <- streamMessage

			Eventually(func() int { return httpmock.GetCallCountInfo()[fmt.Sprintf("POST %s", webhookUrl)] }).Should(Equal(1))
			Eventually(handlerCall).Should(Receive(Equal("success_called")))
		})
	})

	When("receives a recoverable error response from webhook server", func() {
		It("should call the message's OnRecoverableFailureHandler", func() {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			webhookUrl := "/operator-webhook"
			recoverableResponseCode := http.StatusInternalServerError
			httpmock.RegisterResponder(http.MethodPost, webhookUrl, httpmock.NewStringResponder(recoverableResponseCode, ""))

			in := make(chan StreamMessage[*common.RunCompletionEventData])
			webhookSink := &WebhookSink{client: client, operatorWebhook: webhookUrl, in: in}

			go webhookSink.SendEvents(ctx, webhookSink)

			streamMessage := StreamMessage[*common.RunCompletionEventData]{
				Message:            &runCompletionEventData,
				OnCompleteHandlers: onCompHandlers,
			}
			in <- streamMessage

			Eventually(func() int { return httpmock.GetCallCountInfo()[fmt.Sprintf("%s %s", http.MethodPost, webhookUrl)] }).Should(Equal(1))
			Eventually(handlerCall).Should(Receive(Equal("recoverable_failure_called")))
		})
	})

	When("receives an unrecoverable error response from the webhook server", func() {
		It("should call the message's OnUnrecoverableFailureHandler", func() {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			webhookUrl := "/operator-webhook"
			unrecoverableResponseCode := http.StatusGone
			httpmock.RegisterResponder(http.MethodPost, webhookUrl, httpmock.NewStringResponder(unrecoverableResponseCode, ""))

			in := make(chan StreamMessage[*common.RunCompletionEventData])
			webhookSink := &WebhookSink{client: client, operatorWebhook: webhookUrl, in: in}

			go webhookSink.SendEvents(ctx, webhookSink)

			streamMessage := StreamMessage[*common.RunCompletionEventData]{
				Message:            &runCompletionEventData,
				OnCompleteHandlers: onCompHandlers,
			}
			in <- streamMessage

			Eventually(func() int { return httpmock.GetCallCountInfo()[fmt.Sprintf("%s %s", http.MethodPost, webhookUrl)] }).Should(Equal(1))
			Eventually(handlerCall).Should(Receive(Equal("unrecoverable_failure_called")))
		})
	})
})
