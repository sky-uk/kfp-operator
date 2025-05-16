//go:build unit

package server

import (
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
)

var _ = Context("MetricsServer", func() {
	var (
		counter prometheus.Counter
		s       *httptest.Server
	)

	BeforeEach(func() {
		counter = prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "test_metric_total",
				Help: "A test metric",
			},
		)
		prometheus.MustRegister(counter)

		handler := NewMetricsServer(":0").Handler
		s = httptest.NewServer(handler)
	})

	AfterEach(func() {
		s.Close()
	})

	When("metrics endpoint is called after registered metric is updated", func() {
		It("should return the updated metric", func() {
			counter.Inc()
			resp, err := http.Get(s.URL + "/metrics")
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			Expect(string(body)).To(ContainSubstring("test_metric_total 1"))
		})
	})
})
