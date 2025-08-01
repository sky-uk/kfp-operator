//go:build unit

package metrics

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
)

func TestMeterProvider(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MeterProvider Suite")
}

var _ = Describe("InitMeterProvider", func() {
	var (
		serviceName           = "test-service"
		originalMeterProvider = otel.GetMeterProvider()
	)

	AfterEach(func() {
		otel.SetMeterProvider(originalMeterProvider)
	})

	Context("when called with valid parameters", func() {
		It("should set the global meter provider", func() {
			meterProvider, err := InitMeterProvider(serviceName)

			Expect(err).ToNot(HaveOccurred())
			Expect(otel.GetMeterProvider()).To(Equal(meterProvider))
		})

		It("should create a meter provider with custom prometheus exporter options", func() {
			customRegistry := prometheus.NewRegistry()

			customOptions := []promexporter.Option{
				promexporter.WithoutUnits(),
				promexporter.WithoutScopeInfo(),
				promexporter.WithRegisterer(customRegistry),
			}

			meterProvider, err := InitMeterProvider(serviceName, customOptions...)

			Expect(err).ToNot(HaveOccurred())
			Expect(meterProvider).ToNot(BeNil())

			metricFamilies, err := customRegistry.Gather()
			Expect(err).ToNot(HaveOccurred())
			Expect(metricFamilies).ToNot(BeEmpty())
		})
	})
})
