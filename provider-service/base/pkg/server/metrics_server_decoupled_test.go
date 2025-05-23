//go:build decoupled

package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
)

type MetricsTestContext struct {
	Addr     net.Addr
	shutdown func()
}

func NewMetricsTestContext() *MetricsTestContext {
	var ctx = logr.NewContext(context.Background(), logr.Discard())

	listener, err := net.Listen("tcp", ":0")
	Expect(err).NotTo(HaveOccurred())

	shutdown, err := initialiseMetricsServerFromListener(ctx, listener, "stub")
	Expect(err).NotTo(HaveOccurred())

	metricsAddr := listener.Addr()

	return &MetricsTestContext{Addr: metricsAddr, shutdown: shutdown}
}

func TestMetricsServerDecoupledSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metrics Server decoupled Suite")
}

func (mtc MetricsTestContext) getMetrics() string {
	resp, err := http.Get(fmt.Sprintf("http://%s/metrics", mtc.Addr.String()))
	Expect(err).NotTo(HaveOccurred())

	body, err := io.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())

	err = resp.Body.Close()
	Expect(err).NotTo(HaveOccurred())

	return string(body)
}

var _ = Context("Metrics Server", func() {
	When("initialising the metrics server", func() {
		It("returns an error when port is not specified", func() {
			ctx := context.Background()
			err := MetricsServer{}.Start(ctx, config.MetricsConfig{}, "stub")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("metrics.Port must be specified"))
		})

		It("provides metrics in Prometheus format", func() {
			testCtx := NewMetricsTestContext()
			defer testCtx.shutdown()
			Expect(testCtx.getMetrics()).To(MatchRegexp(`(?m)^go_info{version="[\w.]+"} \d$`))
		})
	})
})
