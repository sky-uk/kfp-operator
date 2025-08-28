//go:build decoupled

package sinks

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-resty/resty/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/webhook"
	"github.com/sky-uk/kfp-operator/internal/log"
	"github.com/sky-uk/kfp-operator/pkg/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

func (m StubRCEHandler) Handle(_ context.Context, _ common.RunCompletionEvent) webhook.EventError {
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
	logger, err := log.NewLogger(zapcore.InfoLevel)
	Expect(err).ToNot(HaveOccurred())
	ctx, cancel := createContextWithLogger(logger)

	port, err := randomFreePort()
	Expect(err).ToNot(HaveOccurred())

	scheme := runtime.NewScheme()
	groupVersion := schema.GroupVersion{Group: "pipelines.kubeflow.org", Version: "v1beta1"}
	scheme.AddKnownTypes(groupVersion, &pipelineshub.RunConfiguration{})
	metav1.AddToGroupVersion(scheme, groupVersion)

	rc := pipelineshub.RandomRunConfiguration(common.RandomNamespacedName())

	var handlers []webhook.RunCompletionEventHandler
	stubHandler := StubRCEHandler{}

	httpClient := resty.New()

	BeforeAll(func() {
		handlers = append(handlers, stubHandler)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(rc).Build()
		rcf := webhook.NewRunCompletionFeed(
			fakeClient,
			handlers,
		)

		go func() {
			http.HandleFunc("/events", rcf.HandleEvent(ctx))

			http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
		}()
	})

	AfterAll(func() {
		cancel()
	})

	var handlerCall chan any

	BeforeEach(func() {
		handlerCall = make(chan any)
	})

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

	When("message is valid and successfully sends to the webhook", func() {
		It("should call its OnSuccessHandler", func() {
			inChan := make(chan StreamMessage[*common.RunCompletionEventData])

			_ = NewWebhookSink(ctx, httpClient, fmt.Sprintf("http://localhost:%d/events", port), inChan)

			runCompletionEventData := webhook.RandomRunCompletionEventData()
			runCompletionEventData.RunConfigurationName = &common.NamespacedName{
				Name:      rc.Name,
				Namespace: rc.Namespace,
			}
			runCompletionEventData.RunName = nil

			stm := StreamMessage[*common.RunCompletionEventData]{
				Message:            &runCompletionEventData,
				OnCompleteHandlers: onCompHandlers,
			}

			// send data to channel which should be picked up by sendEvents in webhookSink
			inChan <- stm

			Eventually(handlerCall).Should(Receive(Equal("success_called")))
		})
	})

	When("message is invalid", func() {
		It("should call its OnRecoverableFailureHandler", func() {
			inChan := make(chan StreamMessage[*common.RunCompletionEventData])

			_ = NewWebhookSink(ctx, httpClient, fmt.Sprintf("http://localhost:%d/events", port), inChan)

			emptyRunCompletionData := common.RunCompletionEventData{}

			stm := StreamMessage[*common.RunCompletionEventData]{
				Message:            &emptyRunCompletionData,
				OnCompleteHandlers: onCompHandlers,
			}

			// send data to channel which should be picked up by sendEvents in webhookSink
			inChan <- stm

			Eventually(handlerCall).Should(Receive(Equal("recoverable_failure_called")))
		})
	})

	When("message contains a resource that doesn't exist", func() {
		It("should call its OnUnrecoverableFailureHandler", func() {
			inChan := make(chan StreamMessage[*common.RunCompletionEventData])

			_ = NewWebhookSink(ctx, httpClient, fmt.Sprintf("http://localhost:%d/events", port), inChan)

			runCompletionEventData := webhook.RandomRunCompletionEventData()
			runCompletionEventData.RunConfigurationName = &common.NamespacedName{
				Name:      "doesnt",
				Namespace: "exist",
			}
			runCompletionEventData.RunName = nil

			stm := StreamMessage[*common.RunCompletionEventData]{
				Message:            &runCompletionEventData,
				OnCompleteHandlers: onCompHandlers,
			}

			// send data to channel which should be picked up by sendEvents in webhookSink
			inChan <- stm

			Eventually(handlerCall).Should(Receive(Equal("unrecoverable_failure_called")))
		})
	})
})
