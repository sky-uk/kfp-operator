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
	"net/http"
	"net/http/httptest"
)

var mockUpstreamServiceCallCounter = 0

var _ = BeforeEach(func() {
	mockUpstreamServiceCallCounter = 0
})

type MockUpstreamService struct {
	expectedBody string
}

func (m MockUpstreamService) call(_ context.Context, event common.RunCompletionEvent) error {
	mockUpstreamServiceCallCounter++
	passedBodyBytes, err := json.Marshal(event)
	Expect(err).NotTo(HaveOccurred())
	passedBodyStr := string(passedBodyBytes)
	if m.expectedBody == "error" {
		return errors.New("upstream service error")
	} else if passedBodyStr != m.expectedBody {
		Fail(fmt.Sprintf("Body passed to upstream service does not match expected body, passed - [%s], expected - [%s]", passedBodyStr, m.expectedBody))
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

	noUpstreams := RunCompletionFeed{
		ctx:       ctx,
		upstreams: nil,
	}

	rced := RandomRunCompletionEventData()
	requestBody := DataWrapper{
		Data: rced,
	}
	requestStr, err := json.Marshal(requestBody)
	Expect(err).NotTo(HaveOccurred())

	rce := RandomRunCompletionEventData().ToRunCompletionEvent()
	expectedUpstreamStr, err := json.Marshal(rce)
	Expect(err).NotTo(HaveOccurred())

	upstreams := []UpstreamService{MockUpstreamService{expectedBody: string(expectedUpstreamStr)}, MockUpstreamService{expectedBody: string(expectedUpstreamStr)}}

	eventProcessor := StubbedEventProcessor{
		expectedRunCompletionEventData: &rced,
		returnedRunCompletionEvent:     &rce,
	}

	withUpstreams := RunCompletionFeed{
		ctx:            ctx,
		eventProcessor: eventProcessor,
		upstreams:      upstreams,
	}

	When("called with a valid request", func() {
		It("calls out to configured upstreams passing expected data", func() {
			req, resp := setupRequestResponse(ctx, http.MethodPost, bytes.NewReader(requestStr), HttpContentTypeJSON)

			withUpstreams.handleEvent(resp, req)

			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(mockUpstreamServiceCallCounter).To(Equal(len(upstreams)))
		})
	})

	When("called with empty body", func() {
		It("returns bad request", func() {
			req, resp := setupRequestResponse(ctx, http.MethodPost, nil, HttpContentTypeJSON)

			noUpstreams.handleEvent(resp, req)

			Expect(resp.Code).To(Equal(http.StatusBadRequest))
		})
	})

	When("called with an incorrect content type", func() {
		It("returns unsupported mediatype", func() {
			req, resp := setupRequestResponse(ctx, http.MethodPost, bytes.NewReader(requestStr), "application/xml")

			noUpstreams.handleEvent(resp, req)

			Expect(resp.Code).To(Equal(http.StatusUnsupportedMediaType))
		})
	})

	When("called with an invalid http method", func() {
		It("returns method not allowed error", func() {
			req, resp := setupRequestResponse(ctx, http.MethodGet, bytes.NewReader(requestStr), HttpContentTypeJSON)

			noUpstreams.handleEvent(resp, req)

			Expect(resp.Code).To(Equal(http.StatusMethodNotAllowed))
		})
	})

	When("a upstream returns an error", func() {
		It("returns internal server error", func() {
			upstreams := []UpstreamService{MockUpstreamService{expectedBody: "error"}, MockUpstreamService{expectedBody: string(expectedUpstreamStr)}}
			withErrorUpstream := RunCompletionFeed{
				ctx:            ctx,
				eventProcessor: eventProcessor,
				upstreams:      upstreams,
			}

			req, resp := setupRequestResponse(ctx, http.MethodPost, bytes.NewReader(requestStr), HttpContentTypeJSON)

			withErrorUpstream.handleEvent(resp, req)

			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
			Expect(resp.Body.String()).To(Equal("upstream service error\n"))
			Expect(mockUpstreamServiceCallCounter).To(Equal(1))
		})
	})
})
