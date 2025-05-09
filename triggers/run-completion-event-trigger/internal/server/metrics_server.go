package server

import (
	"context"
	"errors"
	"net"
	"net/http"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type MetricsServer struct{}

func (ms MetricsServer) Start(
	ctx context.Context,
	host string,
	port string,
	reg prom.Gatherer,
) error {

	if port == "" {
		return errors.New("metrics.Port must be specified")
	}

	go serveMetrics(ctx, host, port, reg)

	return nil
}

func serveMetrics(
	ctx context.Context,
	host string,
	port string,
	reg prom.Gatherer,
) {
	logger := common.LoggerFromContext(ctx)
	route := "/metrics"

	addr := net.JoinHostPort(host, port)

	logger.Info("Starting metrics server", "addr", addr, "route", route)

	http.Handle(
		route,
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: true}),
	)

	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Error(err, "Metrics serving failed")
	}
}
