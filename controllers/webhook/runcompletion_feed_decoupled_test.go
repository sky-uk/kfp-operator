//go:build decoupled

package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var mockRCEHandlerHandleCounter = 0

var _ = BeforeEach(func() {
	mockRCEHandlerHandleCounter = 0
})

type MockRCEHandler struct {
	expectedBody string
}

func (m MockRCEHandler) Handle(event common.RunCompletionEvent) EventError {
	mockRCEHandlerHandleCounter++
	passedBodyBytes, err := json.Marshal(event)
	Expect(err).NotTo(HaveOccurred())
	passedBodyStr := string(passedBodyBytes)
	if m.expectedBody == "error" {
		return &FatalError{Msg: "run completion event handler error"}
	} else if m.expectedBody == "not found" {
		return &MissingResourceError{Msg: "resource not found"}
	} else if passedBodyStr != m.expectedBody {
		Fail(
			fmt.Sprintf(
				"Body passed to run completion event handler does not match expected body, passed - [%s], expected - [%s]",
				passedBodyStr,
				m.expectedBody,
			),
		)
	}
	return nil
}

func schemeWithCRDs() *runtime.Scheme {
	scheme := runtime.NewScheme()

	groupVersion := schema.GroupVersion{Group: "pipelines.kubeflow.org", Version: "v1beta1"}
	scheme.AddKnownTypes(groupVersion, &pipelineshub.RunConfiguration{}, &pipelineshub.Run{})

	metav1.AddToGroupVersion(scheme, groupVersion)
	return scheme
}

func setupRequestResponse(ctx context.Context, method string, body io.Reader, contentType string) (*http.Request, *httptest.ResponseRecorder) {
	req, err := http.NewRequestWithContext(ctx, method, "http://example.com/events", body)
	Expect(err).NotTo(HaveOccurred())
	req.Header.Add(HttpHeaderContentType, contentType)
	return req, httptest.NewRecorder()
}

var _ = Describe("Run the run completion feed webhook", Serial, func() {
	ctx := logr.NewContext(context.Background(), logr.Discard())
	rc := pipelineshub.RandomRunConfiguration(common.RandomNamespacedName())
	fakeClient := fake.NewClientBuilder().WithScheme(schemeWithCRDs()).WithObjects(rc).Build()

	noHandlers := RunCompletionFeed{
		ctx:           ctx,
		client:        fakeClient,
		eventHandlers: nil,
	}

	rced := RandomRunCompletionEventData()
	rced.RunConfigurationName = &common.NamespacedName{
		Name:      rc.Name,
		Namespace: rc.Namespace,
	}
	rced.RunName = nil
	requestStr, err := json.Marshal(rced)
	Expect(err).NotTo(HaveOccurred())

	rce := RandomRunCompletionEventData().ToRunCompletionEvent()
	expectedRCEHStr, err := json.Marshal(rce)
	Expect(err).NotTo(HaveOccurred())

	handlers := []RunCompletionEventHandler{
		MockRCEHandler{expectedBody: string(expectedRCEHStr)},
		MockRCEHandler{expectedBody: string(expectedRCEHStr)},
	}

	eventProcessor := StubbedEventProcessor{
		expectedRunCompletionEventData: &rced,
		returnedRunCompletionEvent:     &rce,
	}

	withHandlers := RunCompletionFeed{
		ctx:            ctx,
		client:         fakeClient,
		eventProcessor: eventProcessor,
		eventHandlers:  handlers,
	}

	When("called with a valid request", func() {
		It("calls out to configured run completion event handlers passing expected data", func() {
			req, resp := setupRequestResponse(ctx, http.MethodPost, bytes.NewReader(requestStr), HttpContentTypeJSON)

			withHandlers.handleEvent(resp, req)

			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(mockRCEHandlerHandleCounter).To(Equal(len(handlers)))
		})
	})

	When("called with empty body", func() {
		It("returns bad request", func() {
			req, resp := setupRequestResponse(ctx, http.MethodPost, nil, HttpContentTypeJSON)

			noHandlers.handleEvent(resp, req)

			Expect(resp.Code).To(Equal(http.StatusBadRequest))
		})
	})

	When("called with an incorrect content type", func() {
		It("returns unsupported mediatype", func() {
			req, resp := setupRequestResponse(ctx, http.MethodPost, bytes.NewReader(requestStr), "application/xml")

			noHandlers.handleEvent(resp, req)

			Expect(resp.Code).To(Equal(http.StatusUnsupportedMediaType))
		})
	})

	When("called with an invalid http method", func() {
		It("returns method not allowed error", func() {
			req, resp := setupRequestResponse(ctx, http.MethodGet, bytes.NewReader(requestStr), HttpContentTypeJSON)

			noHandlers.handleEvent(resp, req)

			Expect(resp.Code).To(Equal(http.StatusMethodNotAllowed))
		})
	})

	When("a run completion event handler returns a `Not Found` error", func() {
		It("returns `Gone` server error", func() {
			handlers := []RunCompletionEventHandler{
				MockRCEHandler{expectedBody: "not found"},
			}
			withErrorHandler := RunCompletionFeed{
				ctx:            ctx,
				client:         fakeClient,
				eventProcessor: eventProcessor,
				eventHandlers:  handlers,
			}

			req, resp := setupRequestResponse(ctx, http.MethodPost, bytes.NewReader(requestStr), HttpContentTypeJSON)

			withErrorHandler.handleEvent(resp, req)

			Expect(resp.Code).To(Equal(http.StatusGone))
		})
	})

	When("a run completion event handler returns an error", func() {
		It("returns internal server error", func() {
			handlers := []RunCompletionEventHandler{
				MockRCEHandler{expectedBody: "error"},
				MockRCEHandler{expectedBody: string(expectedRCEHStr)},
			}
			withErrorHandler := RunCompletionFeed{
				ctx:            ctx,
				client:         fakeClient,
				eventProcessor: eventProcessor,
				eventHandlers:  handlers,
			}

			req, resp := setupRequestResponse(ctx, http.MethodPost, bytes.NewReader(requestStr), HttpContentTypeJSON)

			withErrorHandler.handleEvent(resp, req)

			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
			Expect(resp.Body.String()).To(Equal("run completion event handler error\n"))
			Expect(mockRCEHandlerHandleCounter).To(Equal(1))
		})
	})
})
