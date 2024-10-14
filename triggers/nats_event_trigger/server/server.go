package nats_event_trigger

import (
	"context"
	"encoding/json"
	"google.golang.org/protobuf/types/known/emptypb"

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

func (s *Server) ProcessEventFeed(_ context.Context, in *pb.RunCompletionEvent) (*emptypb.Empty, error) {
	eventData, err := json.Marshal(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "marshalling provided event failed")
	}

	err = s.Publisher.Publish(eventData)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to publish event")
	}

	return &emptypb.Empty{}, nil
}
