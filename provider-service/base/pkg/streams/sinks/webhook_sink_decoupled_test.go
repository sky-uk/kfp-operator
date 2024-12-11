//go:build decoupled

package sinks

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/go-resty/resty/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/controllers/webhook"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
	"time"
)

func TestSinksDecoupledSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sinks decoupled Suite")
}

func createContextWithLogger(logger logr.Logger) (ctx context.Context, cancel context.CancelFunc) {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), time.Duration(5000)*time.Millisecond)
	ctxWithLogger := logr.NewContext(ctxWithTimeout, logger)
	return ctxWithLogger, cancel
}

type StubRCEHandler struct {
}

func (m StubRCEHandler) Handle(_ context.Context, _ common.RunCompletionEvent) error {
	return nil
}

func randomFreePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}

	// Extract the port number
	addr := listener.Addr().(*net.TCPAddr)
	port := addr.Port
	err = listener.Close()
	if err != nil {
		return 0, err
	}
	return port, nil
}

var _ = Context("Webhook Sink", Ordered, func() {
	logger, err := common.NewLogger(zapcore.InfoLevel)
	Expect(err).ToNot(HaveOccurred())
	ctx, cancel := createContextWithLogger(logger)

	port, err := randomFreePort()
	Expect(err).ToNot(HaveOccurred())

	scheme := runtime.NewScheme()
	groupVersion := schema.GroupVersion{Group: "pipelines.kubeflow.org", Version: "v1alpha6"}
	scheme.AddKnownTypes(groupVersion, &pipelinesv1.RunConfiguration{})
	metav1.AddToGroupVersion(scheme, groupVersion)

	rc := pipelinesv1.RandomRunConfiguration(apis.RandomLowercaseString())

	var handlers []webhook.RunCompletionEventHandler
	stubHandler := StubRCEHandler{}

	httpClient := resty.New()

	BeforeAll(func() {
		handlers = append(handlers, stubHandler)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(rc).
			Build()
		rcf := webhook.NewRunCompletionFeed(
			ctx,
			fakeClient,
			handlers,
		)
		go func() {
			err = rcf.Start(port)
			if err != nil {
				logger.Error(err, "problem starting run completion feed")
				os.Exit(1)
			}
		}()
	})

	AfterAll(func() {
		cancel()
	})

	handlerCall := make(chan any)

	BeforeEach(func() {
		handlerCall = make(chan any)
	})

	onCompHandlers := OnCompleteHandlers{
		OnSuccessHandler: func() {
			handlerCall <- "success_called"
		},
		OnFailureHandler: func() {
			handlerCall <- "failure_called"
		},
	}

	When("it is valid", func() {
		It("should get success response from webhook", func() {
			inChan := make(chan StreamMessage[*common.RunCompletionEventData])

			_ = NewWebhookSink(ctx, httpClient, fmt.Sprintf("http://localhost:%d/events", port), inChan)

			runCompletionEventData := webhook.RandomRunCompletionEventData()
			runCompletionEventData.RunConfigurationName = &common.NamespacedName{
				Name:      rc.Name,
				Namespace: rc.Namespace,
			}

			stm := StreamMessage[*common.RunCompletionEventData]{
				Message:            &runCompletionEventData,
				OnCompleteHandlers: onCompHandlers,
			}

			// send data to channel which should be picked up by sendEvents in webhookSink
			inChan <- stm

			Eventually(handlerCall).Should(Receive(Equal("success_called")))
		})
	})

	When("it is invalid", func() {
		It("should send request to webhook", func() {
			inChan := make(chan StreamMessage[*common.RunCompletionEventData])

			_ = NewWebhookSink(ctx, httpClient, fmt.Sprintf("http://localhost:%d/events", port), inChan)

			emptyRunCompletionData := common.RunCompletionEventData{}

			stm := StreamMessage[*common.RunCompletionEventData]{
				Message:            &emptyRunCompletionData,
				OnCompleteHandlers: onCompHandlers,
			}

			// send data to channel which should be picked up by sendEvents in webhookSink
			inChan <- stm

			Eventually(handlerCall).Should(Receive(Equal("failure_called")))
		})
	})
})