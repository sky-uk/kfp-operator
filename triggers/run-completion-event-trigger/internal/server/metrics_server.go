package server

import (
	"context"
	"net/http"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type MetricsServer struct{}

func (ms MetricsServer) Start(
	ctx context.Context,
	addr string,
	reg prom.Gatherer,
) error {
	go serveMetrics(ctx, addr, reg)

	return nil
}

func serveMetrics(
	ctx context.Context,
	addr string,
	reg prom.Gatherer,
) {
	logger := common.LoggerFromContext(ctx)
	route := "/metrics"

	logger.Info("Starting metrics server", "addr", addr, "route", route)

	http.Handle(
		route,
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: true}),
	)

	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Error(err, "Metrics serving failed")
	}
}
