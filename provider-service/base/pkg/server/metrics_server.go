package server

import (
	"fmt"
	"net/http"

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
