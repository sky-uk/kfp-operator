package publisher

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
)

var _ = Context("SendEvents", func() {
	var logger, _ = common.NewLogger(zapcore.DebugLevel)
	var ctx = logr.NewContext(context.Background(), logger)

	When("webhook sink receives StreamMessage", func() {
		It("sends RunCompletionEventData to the webhook successfully", func() {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			webhookUrl := "/operator-webhook"
			httpmock.RegisterResponder("POST", webhookUrl, httpmock.NewStringResponder(200, ""))

			in := make(chan any)
			httpWebhook := &HttpWebhookSink{context: ctx, client: client, operatorWebhook: webhookUrl, in: in}

			onSuccessCalled := make(chan any, 1)

			go httpWebhook.SendEvents()
			handlers := OnCompleteHandlers{
				OnSuccessHandler: func() {
					onSuccessCalled <- "ring ring!"
				},
			}
			streamMessage := StreamMessage{OnCompleteHandlers: handlers}
			in <- streamMessage

			Eventually(func() int { return httpmock.GetCallCountInfo()[fmt.Sprintf("POST %s", webhookUrl)] }).Should(Equal(1))
			Eventually(onSuccessCalled).Should(Receive(Equal("ring ring!")))
		})

		It("fails to send RunCompletionEventData to the webhook", func() {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			webhookUrl := "/operator-webhook"
			httpmock.RegisterResponder("POST", webhookUrl, httpmock.NewStringResponder(500, ""))

			in := make(chan any)
			httpWebhook := &HttpWebhookSink{context: ctx, client: client, operatorWebhook: webhookUrl, in: in}

			onFailureCalled := make(chan any)

			go httpWebhook.SendEvents()
			handlers := OnCompleteHandlers{
				OnFailureHandler: func() {
					onFailureCalled <- "ring ring!"
				},
			}
			streamMessage := StreamMessage{OnCompleteHandlers: handlers}
			in <- streamMessage

			Eventually(func() int { return httpmock.GetCallCountInfo()[fmt.Sprintf("POST %s", webhookUrl)] }).Should(Equal(1))
			Eventually(onFailureCalled).Should(Receive(Equal("ring ring!")))
		})
	})
})
