//go:build decoupled

package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	"io"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/http"
	"net/http/httptest"
)

var mockRCEHandlerHandleCounter = 0

var _ = BeforeEach(func() {
	mockRCEHandlerHandleCounter = 0
})

type MockRCEHandler struct {
	expectedBody string
}

func (m MockRCEHandler) Handle(event common.RunCompletionEvent) error {
	mockRCEHandlerHandleCounter++
	passedBodyBytes, err := json.Marshal(event)
	Expect(err).NotTo(HaveOccurred())
	passedBodyStr := string(passedBodyBytes)
	if m.expectedBody == "error" {
		return errors.New("run completion event handler error")
	} else if m.expectedBody == "not found" {
		return k8sErrors.NewNotFound(schema.GroupResource{}, "run completion event handler not found")
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

func setupRequestResponse(ctx context.Context, method string, body io.Reader, contentType string) (*http.Request, *httptest.ResponseRecorder) {
	req, err := http.NewRequestWithContext(ctx, method, "http://example.com/events", body)
	Expect(err).NotTo(HaveOccurred())
	req.Header.Add(HttpHeaderContentType, contentType)
	return req, httptest.NewRecorder()
}

var _ = Describe("Run the run completion feed webhook", Serial, func() {
	ctx := logr.NewContext(context.Background(), logr.Discard())

	noHandlers := RunCompletionFeed{
		ctx:           ctx,
		eventHandlers: nil,
	}

	rced := RandomRunCompletionEventData()
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
