package server

import (
	"context"
	"errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"

	configLoader "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/cmd/config"
	"github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/internal/converters"
	"github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/internal/publisher"
	pb "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedRunCompletionEventTriggerServer
	Config    *configLoader.Config
	Publisher publisher.PublisherHandler
}

func (s *Server) ProcessEventFeed(_ context.Context, runCompletion *pb.RunCompletionEvent) (*emptypb.Empty, error) {

	commonRunCompletionEvent, err := converters.ProtoRunCompletionToCommon(runCompletion)
	if err != nil {
		err = status.Error(codes.InvalidArgument, "Proto run completion event could not be converted to a common run completion event.")
		return nil, err
	}

	if err = s.Publisher.Publish(commonRunCompletionEvent); err != nil {
		var marshallingError *publisher.MarshallingError
		var connectionError *publisher.ConnectionError
		switch {
		case errors.As(err, &marshallingError):
			return nil, status.Error(codes.InvalidArgument, "failed to marshal event")
		case errors.As(err, &connectionError):
			return nil, status.Error(codes.Internal, "publisher request to upstream failed")
		default:
			return nil, status.Error(codes.Internal, "unexpected error occurred")
		}
	}

	log.Printf("Run Completion Event Processed for %s", commonRunCompletionEvent.RunId)

	return &emptypb.Empty{}, nil
}
