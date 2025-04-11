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
	logger := common.LoggerFromContext(ctx).WithValues("runId", runCompletion.RunId)

	commonRunCompletionEvent, err := converters.ProtoRunCompletionToCommon(runCompletion)
	if err != nil {
		errMsg := "Proto run completion event could not be converted to a common run completion event."
		logger.Error(err, errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	if err = s.Publisher.Publish(commonRunCompletionEvent); err != nil {
		var marshallingError *publisher.MarshallingError
		var connectionError *publisher.ConnectionError

		switch {
		case errors.As(err, &marshallingError):
			errMsg := "failed to marshal event"
			logger.Error(err, errMsg)
			return nil, status.Error(codes.InvalidArgument, errMsg)
		case errors.As(err, &connectionError):
			errMsg := "publisher request to upstream failed"
			logger.Error(err, errMsg)
			return nil, status.Error(codes.Internal, errMsg)
		default:
			errMsg := "unexpected error occurred"
			logger.Error(err, errMsg)
			return nil, status.Error(codes.Internal, errMsg)
		}
	}

	logger.Info("Run Completion Event Processed")

	return &emptypb.Empty{}, nil
}
