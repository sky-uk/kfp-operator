package metrics

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func InitMeterProvider(serviceName string, exporterOptions ...prometheus.Option) (*metric.MeterProvider, error) {
	exporter, err := prometheus.New(exporterOptions...)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(newResource(serviceName)),
		metric.WithReader(exporter),
	)

	otel.SetMeterProvider(meterProvider)

	return meterProvider, nil
}

func newResource(serviceName string) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
	)
}
