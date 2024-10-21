package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	configLoader "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/cmd/config"
	"github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/internal/publisher"
	"github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/internal/server"
	pb "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/proto"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	config, err := configLoader.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config on startup %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", config.ServerConfig.Host, config.ServerConfig.Port))
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", config.ServerConfig.Port, err)
	}

	nc, err := nats.Connect(config.NATSConfig.ServerConfig.ToUrl())
	if err != nil {
		log.Fatalf("failed to connect to NATS server: %v", err)
	}
	defer nc.Close()

	s := grpc.NewServer()

	natsPublisher := publisher.NatsPublisher{
		NatsConn: nc,
		Subject:  config.NATSConfig.Subject,
	}

	pb.RegisterRunCompletionEventTriggerServer(s, &server.Server{Config: config, Publisher: natsPublisher})

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
