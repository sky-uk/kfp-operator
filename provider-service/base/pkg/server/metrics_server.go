package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
)

type MetricsServer struct{}

func (ms MetricsServer) Start(ctx context.Context, cfg config.MetricsConfig, serviceName string) error {
	if cfg.Port == 0 {
		return errors.New("metrics.Port must be specified")
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return err
	}

	onShutdown, err := initialiseMetricsServerFromListener(ctx, listener, serviceName)
	if err != nil {
		return err
	}
	defer onShutdown()
	return nil
}

func initialiseMetricsServerFromListener(ctx context.Context, listener net.Listener, serviceName string) (func(), error) {
	logger := common.LoggerFromContext(ctx)

	meterProvider, err := newMeterProvider(serviceName)
	if err != nil {
		return nil, err
	}
	otel.SetMeterProvider(meterProvider)

	go serveMetrics(ctx, listener)

	return func() {
		if err := meterProvider.Shutdown(ctx); err != nil {
			logger.Error(err, "Error shutting down metrics server")
		}
	}, nil
}

func newResource(serviceName string) *resource.Resource {
	return resource.NewWithAttributes(semconv.SchemaURL,
		semconv.ServiceName(fmt.Sprintf("provider-service-%s", serviceName)),
	)
}

func newMeterProvider(serviceName string) (*metric.MeterProvider, error) {
	metricExporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(metric.WithResource(newResource(serviceName)), metric.WithReader(metricExporter))
	return meterProvider, nil
}

func serveMetrics(ctx context.Context, listener net.Listener) {
	logger := common.LoggerFromContext(ctx)
	route := "/metrics"

	http.Handle(route, promhttp.Handler())
	if err := http.Serve(listener, nil); err != nil {
		logger.Error(err, "Metrics serving failed")
	}
	logger.Info(fmt.Sprintf("Serving metrics from %s at %s", listener.Addr().String(), route))
}
