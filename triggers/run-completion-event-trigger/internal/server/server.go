package server

import (
	"context"
	"errors"
	"github.com/sky-uk/kfp-operator/argo/common"
	configLoader "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/cmd/config"
	"github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/internal/converters"
	"github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/internal/publisher"
	pb "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedRunCompletionEventTriggerServer
	Config    *configLoader.Config
	Publisher publisher.PublisherHandler
}

func (s *Server) ProcessEventFeed(
	ctx context.Context,
	runCompletion *pb.RunCompletionEvent,
) (*emptypb.Empty, error) {
	logger := common.LoggerFromContext(ctx)

	commonRunCompletionEvent, err := converters.ProtoRunCompletionToCommon(runCompletion)
	if err != nil {
		err = status.Error(codes.InvalidArgument, "Proto run completion event could not be converted to a common run completion event.")
		return nil, err
	}

	if err = s.Publisher.Publish(commonRunCompletionEvent); err != nil {
		var marshallingError *publisher.MarshallingError
		var connectionError *publisher.ConnectionError

		logger.Error(err, "Failed to publish event", "runId", runCompletion.RunId)

		switch {
		case errors.As(err, &marshallingError):
			return nil, status.Error(codes.InvalidArgument, "failed to marshal event")
		case errors.As(err, &connectionError):
			return nil, status.Error(codes.Internal, "publisher request to upstream failed")
		default:
			return nil, status.Error(codes.Internal, "unexpected error occurred")
		}
	}

	logger.Info("Run Completion Event Processed", "RunId", commonRunCompletionEvent.RunId)

	return &emptypb.Empty{}, nil
}
