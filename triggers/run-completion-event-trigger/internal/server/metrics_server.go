package server

import (
	"net/http"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewMetricsServer(
	addr string,
	reg prom.Gatherer,
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
