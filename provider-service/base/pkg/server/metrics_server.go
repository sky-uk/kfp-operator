package server

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewMetricsServer(host string, port int) *http.Server {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: newMetricsHandler(),
	}

	return server
}

func newMetricsHandler() http.Handler {
	mux := chi.NewRouter()
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

func initMeterProvider(serviceName string) (*metric.MeterProvider, error) {
	exporter, err := prometheus.New()
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

type MetricsServer struct{}

func newResource(serviceName string) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(fmt.Sprintf("provider-service-%s", serviceName)),
	)
}
