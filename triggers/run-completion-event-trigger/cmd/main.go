package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		channel := make(chan os.Signal, 1)
		signal.Notify(channel, syscall.SIGINT, syscall.SIGTERM)
		<-channel
		logger.Info("Received shutdown signal")
		cancel()
	}()

	config, err := configLoader.LoadConfig()
	if err != nil {
		logger.Error(err, "Failed to load config file on startup")
		panic(err)
	}

	nc, err := nats.Connect(config.NATSConfig.ServerConfig.ToAddr())
	if err != nil {
		logger.Error(err, "failed to connect to NATS server", "addr", config.NATSConfig.ServerConfig.ToAddr())
		panic(err)
	}
	defer nc.Close()

	natsPublisher := publisher.NewNatsPublisher(ctx, nc, config.NATSConfig.Subject)

	srvMetrics := server.NewServerMetrics()

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			srvMetrics.UnaryServerInterceptor(),
			unaryLoggerInterceptor(logger),
		),
	)

	pb.RegisterRunCompletionEventTriggerServer(
		grpcServer,
		&server.Server{Config: config, Publisher: natsPublisher},
	)
	healthpb.RegisterHealthServer(
		grpcServer,
		&server.HealthServer{HealthCheck: natsPublisher},
	)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", config.ServerConfig.ToAddr())
	if err != nil {
		logger.Error(err, "Failed to listen", "port", config.ServerConfig.Port)
		panic(err)
	}

	metricsSrv := server.NewMetricsServer(config.MetricsConfig.ToAddr())

	var waitGroup sync.WaitGroup

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		logger.Info("Starting gRPC server", "addr", config.ServerConfig.ToAddr())
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error(err, "gRPC server exited unexpectedly")
			cancel()
		}
	}()

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		logger.Info("Starting metrics server", "addr", metricsSrv.Addr)
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err, "Metrics server exited unexpectedly")
			cancel()
		}
	}()

	<-ctx.Done()
	logger.Info("Context cancelled, shutting down servers...")

	grpcServer.GracefulStop()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		logger.Error(err, "Failed to shutdown metrics server cleanly")
	}

	waitGroup.Wait()
	logger.Info("Shutdown complete")
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
