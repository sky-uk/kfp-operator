package run_completion_event_trigger

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"

	configLoader "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/config"
	pb "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedRunCompletionEventTriggerServer
	Config    *configLoader.Config
	Publisher PublisherHandler
}

func (s *Server) ProcessEventFeed(_ context.Context, runCompletion *pb.RunCompletionEvent) (*emptypb.Empty, error) {

	commonRunCompletionEvent, err := pb.ProtoRunCompletionToCommon(runCompletion)
	if err != nil {
		err = status.Error(codes.InvalidArgument, "Proto run completion event could not be converted to a common run completion event.")
		return nil, err
	}

	marshalErr, connErr := s.Publisher.Publish(commonRunCompletionEvent)

	if marshalErr != nil {
		err = status.Error(codes.InvalidArgument, "failed to marshal event")
		return nil, err
	}

	if connErr != nil {
		err = status.Error(codes.Internal, "publisher request to upstream failed")
		return nil, err
	}

	log.Printf("Run Completion Event Processed for %s", commonRunCompletionEvent.RunId)

	return &emptypb.Empty{}, nil
}
