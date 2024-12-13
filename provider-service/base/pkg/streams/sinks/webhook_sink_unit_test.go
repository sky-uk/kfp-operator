//go:build unit

package sinks

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	resty "github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	"go.uber.org/zap/zapcore"
	"testing"
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
		OnFailureHandler: func() {
			handlerCall <- "failure_called"
		},
	}

	When("webhook sink receives a valid StreamMessage", func() {
		It("sends RunCompletionEventData to the webhook successfully and triggers the message OnSuccessHandler", func() {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			webhookUrl := "/operator-webhook"
			httpmock.RegisterResponder("POST", webhookUrl, httpmock.NewStringResponder(200, ""))

			in := make(chan StreamMessage[*common.RunCompletionEventData])
			webhookSink := &WebhookSink{context: ctx, client: client, operatorWebhook: webhookUrl, in: in}

			go webhookSink.SendEvents()

			streamMessage := StreamMessage[*common.RunCompletionEventData]{
				Message:            &runCompletionEventData,
				OnCompleteHandlers: onCompHandlers,
			}
			in <- streamMessage

			Eventually(func() int { return httpmock.GetCallCountInfo()[fmt.Sprintf("POST %s", webhookUrl)] }).Should(Equal(1))
			Eventually(handlerCall).Should(Receive(Equal("success_called")))
		})
	})

	When("webhook sink receives an invalid StreamMessage", func() {
		It("should call its `OnFailureHandler` function", func() {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			webhookUrl := "/operator-webhook"
			someNon200ResponseCode := 500
			httpmock.RegisterResponder("POST", webhookUrl, httpmock.NewStringResponder(someNon200ResponseCode, ""))

			in := make(chan StreamMessage[*common.RunCompletionEventData])
			webhookSink := &WebhookSink{context: ctx, client: client, operatorWebhook: webhookUrl, in: in}

			go webhookSink.SendEvents()

			streamMessage := StreamMessage[*common.RunCompletionEventData]{
				Message:            &runCompletionEventData,
				OnCompleteHandlers: onCompHandlers,
			}
			in <- streamMessage

			Eventually(func() int { return httpmock.GetCallCountInfo()[fmt.Sprintf("POST %s", webhookUrl)] }).Should(Equal(1))
			Eventually(handlerCall).Should(Receive(Equal("failure_called")))
		})
	})
})
