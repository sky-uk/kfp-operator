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
		reg     *prometheus.Registry
		s       *httptest.Server
	)

	BeforeEach(func() {
		reg = prometheus.NewRegistry()

		counter = prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "test_metric_total",
				Help: "A test metric",
			},
		)
		Expect(reg.Register(counter)).To(Succeed())

		handler := NewMetricsServer(":0", reg).Handler
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
