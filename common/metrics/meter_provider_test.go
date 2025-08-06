//go:build unit

package metrics

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
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
				promexporter.WithRegisterer(customRegistry),
			}

			meterProvider, err := InitMeterProvider(serviceName, customOptions...)

			Expect(err).ToNot(HaveOccurred())
			Expect(meterProvider).ToNot(BeNil())

			meter := meterProvider.Meter("test-meter")
			counter, err := meter.Int64Counter(
				"test_counter",
				metric.WithDescription("Test counter for validation"),
			)
			Expect(err).ToNot(HaveOccurred())

			counter.Add(context.Background(), 42, metric.WithAttributes(attribute.String("test", "value")))

			metricFamilies, err := customRegistry.Gather()
			Expect(err).ToNot(HaveOccurred())

			foundMetric, found := lo.Find(metricFamilies, func(mf *dto.MetricFamily) bool {
				return mf.GetName() == "test_counter_total"
			})
			Expect(found).To(BeTrue(), "Custom metric should be present in custom registry")
			Expect(foundMetric.GetMetric()).To(HaveLen(1))
			Expect(foundMetric.GetMetric()[0].GetCounter().GetValue()).To(Equal(float64(42)))
		})
	})
})
