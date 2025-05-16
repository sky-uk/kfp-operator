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
) *http.Server {
	return &http.Server{
		Addr: addr,
		Handler: func() http.Handler {
			mux := http.NewServeMux()
			mux.Handle(
				"/metrics",
				promhttp.Handler(),
			)
			return mux
		}(),
	}
}

// NewPrometheusRegistryAndServerMetrics creates new ServerMetrics using
// go-grpc-middleware and registers it to a prometheus registry.
func NewPrometheusRegistryAndServerMetrics() *grpcprom.ServerMetrics {
	srvMetrics := grpcprom.NewServerMetrics()

	prometheus.MustRegister(srvMetrics)

	return srvMetrics
}
