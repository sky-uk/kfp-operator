package server

import (
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

// NewPrometheusRegistryAndServerMetrics creates new ServerMetrics using
// go-grpc-middleware, which has various counters on gRPC in-flight and handled
// requests. The service name and namespace are injected as labels to all the
// counters, and the ServerMetrics are registerd to a prometheus registry.
func NewPrometheusRegistryAndServerMetrics(
	namespace string,
	name string,
) (*prometheus.Registry, *grpcprom.ServerMetrics) {
	labels := mkLabels(name, namespace)
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerCounterOptions(
			grpcprom.WithConstLabels(labels),
		),
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets(
				[]float64{
					0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120,
				},
			),
			grpcprom.WithHistogramConstLabels(labels),
		),
	)

	reg := prometheus.NewRegistry()
	reg.MustRegister(srvMetrics)

	return reg, srvMetrics
}

func mkLabels(namespace string, name string) prometheus.Labels {
	return prometheus.Labels{
		"service.name":      name,
		"service.namespace": namespace,
	}
}
