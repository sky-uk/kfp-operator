package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/sky-uk/kfp-operator/argo/common"
)

type MetricsServer struct{}

func (ms MetricsServer) Start(ctx context.Context, port int, serviceName string) (shutdown func(), err error) {
	if port == 0 {
		return nil, errors.New("metrics.Port must be specified")
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	return initialiseMetricsServerFromListener(ctx, listener, serviceName)
}

func initialiseMetricsServerFromListener(ctx context.Context, listener net.Listener, serviceName string) (func(), error) {
	logger := common.LoggerFromContext(ctx)

	metricExporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(metricExporter),
		metric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(fmt.Sprintf("provider-service-%s", serviceName)),
		)),
	)

	otel.SetMeterProvider(provider)

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		logger.Info(fmt.Sprintf("Serving metrics at http://%s/metrics", listener.Addr().String()))
		if err := http.Serve(listener, mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(err, "Metrics HTTP server failed")
		}
	}()

	return func() {
		if err := provider.Shutdown(ctx); err != nil {
			logger.Error(err, "Error shutting down metrics provider")
		}
	}, nil
}
