//go:build unit

package webhook

import (
	"bytes"
	"context"
	"net/http"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/pkg/common"
	"go.uber.org/zap/zapcore"
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

var _ = Context("extractRunCompletionEventData", func() {
	logger, _ := common.NewLogger(zapcore.DebugLevel)
	ctx := logr.NewContext(context.Background(), logger)
	rcf := RunCompletionFeed{}

	When("passed a valid request", func() {
		It("returns RunCompletionEventData", func() {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://example.com/events", bytes.NewReader([]byte("{\"hello\":\"world\"}")))
			Expect(err).NotTo(HaveOccurred())

			eventData, err := rcf.extractRunCompletionEventData(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			Expect(eventData).To(Equal(&common.RunCompletionEventData{}))
		})
	})

	When("passed a request with empty body", func() {
		It("returns an error", func() {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://example.com/events", bytes.NewReader([]byte("")))
			Expect(err).NotTo(HaveOccurred())
			_, err = rcf.extractRunCompletionEventData(ctx, req)
			Expect(err.Error()).To(Equal("request body is empty"))
		})
	})

	When("passed a request with invalid body", func() {
		It("returns an error", func() {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://example.com/events", bytes.NewReader([]byte("hello world")))
			Expect(err).NotTo(HaveOccurred())
			_, err = rcf.extractRunCompletionEventData(ctx, req)
			Expect(err.Error()).To(Equal("invalid character 'h' looking for beginning of value"))
		})
	})
})
