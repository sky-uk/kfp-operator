//go:build unit

package webhook

import (
	"bytes"
	"context"
	"errors"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/common"
	"go.uber.org/zap/zapcore"
	"net/http"
)

var _ = Context("getRequestBody", func() {
	var logger, _ = common.NewLogger(zapcore.DebugLevel)
	var ctx = logr.NewContext(context.Background(), logger)

	When("valid request", func() {
		It("returns request body contents", func() {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://example.com/events", bytes.NewReader([]byte("hello world")))
			Expect(err).NotTo(HaveOccurred())
			Expect(getRequestBody(ctx, req)).To(Equal([]byte("hello world")))
		})
	})

	When("invalid body passed", func() {
		It("is empty returns an error", func() {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://example.com/events", bytes.NewReader([]byte("")))
			Expect(err).NotTo(HaveOccurred())
			_, err = getRequestBody(ctx, req)
			Expect(err.Error()).To(Equal("request body is empty"))
		})

		It("is nil returns an error", func() {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://example.com/events", nil)
			Expect(err).NotTo(HaveOccurred())
			_, err = getRequestBody(ctx, req)
			Expect(err.Error()).To(Equal("request body is nil"))
		})
	})
})

var _ = Context("extractRunCompletionEvent", func() {
	logger, _ := common.NewLogger(zapcore.DebugLevel)
	ctx := logr.NewContext(context.Background(), logger)
	rcf := RunCompletionFeed{ctx: ctx}

	When("valid request", func() {
		It("returns event data in raw json and headers", func() {
			processor := StubbedEventProcessor{
				returnedRunCompletionEvent: &common.RunCompletionEvent{},
				expectedError:              nil,
			}
			rcf.eventProcessor = processor

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://example.com/events", bytes.NewReader([]byte("{\"hello\":\"world\"}")))
			Expect(err).NotTo(HaveOccurred())

			event, err := rcf.extractRunCompletionEvent(req)
			Expect(err).NotTo(HaveOccurred())

			Expect(event).To(Equal(processor.returnedRunCompletionEvent))
		})

		It("returns error on event processor error", func() {
			processor := StubbedEventProcessor{
				returnedRunCompletionEvent: &common.RunCompletionEvent{},
				expectedError:              errors.New("an error occurred"),
			}
			rcf.eventProcessor = processor

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://example.com/events", bytes.NewReader([]byte("{\"hello\":\"world\"}")))
			Expect(err).NotTo(HaveOccurred())
			_, err = rcf.extractRunCompletionEvent(req)
			Expect(err).To(MatchError("an error occurred"))
		})

		It("returns error on event processor return empty event", func() {
			processor := StubbedEventProcessor{
				returnedRunCompletionEvent: nil,
				expectedError:              nil,
			}
			rcf.eventProcessor = processor

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://example.com/events", bytes.NewReader([]byte("{\"hello\":\"world\"}")))
			Expect(err).NotTo(HaveOccurred())
			_, err = rcf.extractRunCompletionEvent(req)
			Expect(err).To(MatchError("event data is empty"))
		})
	})

	When("empty body passed", func() {
		It("returns an error", func() {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://example.com/events", bytes.NewReader([]byte("")))
			Expect(err).NotTo(HaveOccurred())
			_, err = rcf.extractRunCompletionEvent(req)
			Expect(err.Error()).To(Equal("request body is empty"))
		})
	})

	When("invalid body passed", func() {
		It("returns an error", func() {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://example.com/events", bytes.NewReader([]byte("hello world")))
			Expect(err).NotTo(HaveOccurred())
			_, err = rcf.extractRunCompletionEvent(req)
			Expect(err.Error()).To(Equal("invalid character 'h' looking for beginning of value"))
		})
	})
})
