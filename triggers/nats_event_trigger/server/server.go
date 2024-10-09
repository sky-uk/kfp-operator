package nats_event_trigger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/nats-io/nats.go"
	configLoader "github.com/sky-uk/kfp-operator/triggers/nats_event_trigger/config"
	pb "github.com/sky-uk/kfp-operator/triggers/nats_event_trigger/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	pb.UnimplementedNATSEventTriggerServer
	config         *configLoader.Config
	NATSConnection *nats.Conn
}

func (s *server) ProcessEventFeed(ctx context.Context, in *pb.RunCompletionFeed) (*emptypb.Empty, error) {
	event_data, err := json.Marshal(in)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "failed to connect to marshal event")
	}

	err = s.NATSConnection.Publish(s.config.NATSConfig.Subject, []byte(event_data))
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "failed to publish event")
	}

	return &emptypb.Empty{}, nil
}

func Start() error {
	config, err := configLoader.LoadConfig()
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", config.ServerConfig.Host, config.ServerConfig.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %v", config.ServerConfig.Port, err)
	}

	nc, err := nats.Connect(config.NATSConfig.ServerConfig.ToUrl())

	if err != nil {
		return fmt.Errorf("failed to connect to NATS server: %v", err)
	}

	defer nc.Close()

	s := grpc.NewServer()
	pb.RegisterNATSEventTriggerServer(s, &server{config: config, NATSConnection: nc})
	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	return nil
}
