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
// go-grpc-middleware and registers it to a prometheus registry.
func NewPrometheusRegistryAndServerMetrics() (*prometheus.Registry, *grpcprom.ServerMetrics) {
	srvMetrics := grpcprom.NewServerMetrics()

	reg := prometheus.NewRegistry()
	reg.MustRegister(srvMetrics)

	return reg, srvMetrics
}
