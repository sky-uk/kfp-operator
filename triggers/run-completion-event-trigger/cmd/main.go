package main

import (
	"context"
	"net"

	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
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

	lis, err := net.Listen("tcp", config.ServerConfig.ToAddr())
	if err != nil {
		logger.Error(err, "Failed to listen", "port", config.ServerConfig.Port)
		panic(err)
	}

	nc, err := nats.Connect(config.NATSConfig.ServerConfig.ToAddr())
	if err != nil {
		logger.Error(err, "failed to connect to NATS server", "addr", config.NATSConfig.ServerConfig.ToAddr())
		panic(err)
	}
	defer nc.Close()

	natsPublisher := publisher.NewNatsPublisher(ctx, nc, config.NATSConfig.Subject)

	f := server.ServerMetricz{}
	reg, srvMetrics := f.NewServerMetricz("runcompletioneventtrigger", "fuckknows")

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			srvMetrics.UnaryServerInterceptor(),
			unaryLoggerInterceptor(logger),
		),
	)

	pb.RegisterRunCompletionEventTriggerServer(s, &server.Server{Config: config, Publisher: natsPublisher})

	healthpb.RegisterHealthServer(s, &server.HealthServer{HealthCheck: natsPublisher})

	reflection.Register(s)

	server.MetricsServer{}.Start(ctx, config.MetricsConfig.ToAddr(), reg)

	logger.Info("Listening at", "addr", config.ServerConfig.ToAddr())
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
