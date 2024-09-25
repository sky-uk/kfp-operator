//go:build unit

package webhook

import (
	"bytes"
	"context"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
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
			Expect(getRequestBody(req)).To(Equal([]byte("hello world")))
		})
	})

	When("invalid body passed", func() {
		It("is empty returns an error", func() {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://example.com/events", bytes.NewReader([]byte("")))
			Expect(err).NotTo(HaveOccurred())
			_, err = getRequestBody(req)
			Expect(err.Error()).To(Equal("request body is empty"))
		})

		It("is nil returns an error", func() {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://example.com/events", nil)
			Expect(err).NotTo(HaveOccurred())
			_, err = getRequestBody(req)
			Expect(err.Error()).To(Equal("request body is nil"))
		})
	})
})

var _ = Context("extractEventData", func() {
	var logger, _ = common.NewLogger(zapcore.DebugLevel)
	var ctx = logr.NewContext(context.Background(), logger)

	When("valid request", func() {
		It("returns event data in raw json and headers", func() {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://example.com/events", bytes.NewReader([]byte("hello world")))
			req.Header.Add("hello", "world")
			Expect(err).NotTo(HaveOccurred())
			eventData, err := extractEventData(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(eventData.Body.MarshalJSON()).To(Equal([]byte("hello world")))
			Expect(eventData.Header.Get("hello")).To(Equal("world"))
		})
	})

	When("invalid body passed", func() {
		It("returns an error", func() {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://example.com/events", bytes.NewReader([]byte("")))
			Expect(err).NotTo(HaveOccurred())
			_, err = extractEventData(req)
			Expect(err.Error()).To(Equal("request body is empty"))
		})
	})
})
