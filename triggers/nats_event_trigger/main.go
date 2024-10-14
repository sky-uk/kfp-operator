package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	configLoader "github.com/sky-uk/kfp-operator/triggers/nats_event_trigger/config"
	pb "github.com/sky-uk/kfp-operator/triggers/nats_event_trigger/proto"
	server "github.com/sky-uk/kfp-operator/triggers/nats_event_trigger/server"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	config, err := configLoader.LoadConfig()
	if err != nil {
		log.Fatal(err)
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

	publisher := server.NatsPublisher{
		NatsConn: nc,
		Subject:  config.NATSConfig.Subject,
	}

	pb.RegisterNATSEventTriggerServer(s, &server.Server{Config: config, Publisher: publisher})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	log.Printf("gRPC server listening at %v", lis.Addr())
}
