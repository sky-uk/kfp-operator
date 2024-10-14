package nats_event_trigger

import (
	"context"
	"encoding/json"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"

	configLoader "github.com/sky-uk/kfp-operator/triggers/nats_event_trigger/config"
	pb "github.com/sky-uk/kfp-operator/triggers/nats_event_trigger/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedNATSEventTriggerServer
	Config    *configLoader.Config
	Publisher PublisherHandler
}

func (s *Server) ProcessEventFeed(ctx context.Context, in *pb.RunCompletionEvent) (*emptypb.Empty, error) {
	eventData, err := json.Marshal(in)
	if err != nil {
		err = status.Error(codes.InvalidArgument, "marshalling provided event failed")
		return nil, err
	}

	err = s.Publisher.Publish(eventData)
	if err != nil {
		err = status.Error(codes.Internal, "failed to publish event")
		return nil, err
	}

	log.Printf("Run Completion Event Processed for %s", in.RunId)

	return &emptypb.Empty{}, nil
}
