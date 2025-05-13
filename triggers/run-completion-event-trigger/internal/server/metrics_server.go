package server

import (
	"net/http"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewMetricsServer creates a HTTP server that handles the given metrics
// Gatherer and serves it on the /metrics endpoint
func NewMetricsServer(
	addr string,
	reg prometheus.Gatherer,
) *http.Server {
	return &http.Server{
		Addr: addr,
		Handler: func() http.Handler {
			mux := http.NewServeMux()
			mux.Handle(
				"/metrics",
				promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
			)
			return mux
		}(),
	}
}

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
