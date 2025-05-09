package main

import (
	"context"
	"fmt"
	"net"

	"github.com/go-logr/logr"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sky-uk/kfp-operator/argo/common"
	configLoader "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/cmd/config"
	"github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/internal/publisher"
	"github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/internal/server"
	pb "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/proto"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger, err := common.NewLogger(zapcore.InfoLevel)
	if err != nil {
		logger.Error(err, "Failed to create zap logger")
		panic(err)
	}

	ctx := logr.NewContext(context.Background(), logger)

	config, err := configLoader.LoadConfig()
	if err != nil {
		logger.Error(err, "Failed to load config file on startup")
		panic(err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", config.ServerConfig.Host, config.ServerConfig.Port))
	if err != nil {
		logger.Error(err, "Failed to listen", "port", config.ServerConfig.Port)
		panic(err)
	}

	nc, err := nats.Connect(config.NATSConfig.ServerConfig.ToUrl())
	if err != nil {
		logger.Error(err, "failed to connect to NATS server", "url", config.NATSConfig.ServerConfig.ToUrl())
		panic(err)
	}
	defer nc.Close()

	natsPublisher := publisher.NewNatsPublisher(ctx, nc, config.NATSConfig.Subject)

	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)
	reg := prometheus.NewRegistry()

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			srvMetrics.UnaryServerInterceptor(),
			unaryLoggerInterceptor(logger),
		),
	)
	srvMetrics.InitializeMetrics(s)

	pb.RegisterRunCompletionEventTriggerServer(s, &server.Server{Config: config, Publisher: natsPublisher})

	healthpb.RegisterHealthServer(s, &server.HealthServer{HealthCheck: natsPublisher})

	reflection.Register(s)

	server.MetricsServer{}.Start(ctx, 8081, "runcompletioneventtrigger", reg)

	logger.Info("Listening at", "host", config.ServerConfig.Host, "port", config.ServerConfig.Port)
	if err := s.Serve(lis); err != nil {
		logger.Error(err, "failed to serve grpc service")
		panic(err)
	}

}

func unaryLoggerInterceptor(baseLogger logr.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		ctx = logr.NewContext(ctx, baseLogger)

		return handler(ctx, req)
	}
}
