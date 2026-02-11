package webhook

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sky-uk/kfp-operator/pkg/common"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/sky-uk/kfp-operator/internal/config"
	pb "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/proto"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type RunCompletionEventTrigger struct {
	EndPoint          config.Endpoint
	Client            pb.RunCompletionEventTriggerClient
	ConnectionHandler func() error
}

func NewRunCompletionEventTrigger(ctx context.Context, endpoint config.Endpoint) RunCompletionEventTrigger {
	logger := log.FromContext(ctx)

	conn, err := grpc.NewClient(endpoint.URL(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error(err, "Error creating grpc client")
	}

	logger.Info("RunCompletionEventTrigger client connected to", "endpoint", endpoint.URL())

	client := pb.NewRunCompletionEventTriggerClient(conn)

	return RunCompletionEventTrigger{
		EndPoint: endpoint,
		Client:   client,
		ConnectionHandler: func() error {
			if err := conn.Close(); err != nil {
				return err
			}
			return nil
		},
	}
}

func (rcet RunCompletionEventTrigger) Handle(ctx context.Context, event common.RunCompletionEvent) EventError {
	runCompletionEvent, err := RunCompletionEventToProto(event)
	if err != nil {
		return &InvalidEvent{err.Error()}
	}
	_, err = rcet.Client.ProcessEventFeed(ctx, runCompletionEvent)
	if err != nil {
		return &FatalError{err.Error()}
	}

	return nil
}

func RunCompletionEventToProto(event common.RunCompletionEvent) (*pb.RunCompletionEvent, error) {
	pipelineName, err := event.PipelineName.String()
	if err != nil {
		return nil, fmt.Errorf("failed to format pipeline name for proto run completion event: %w", err)
	}

	var runConfigurationName string
	if event.RunConfigurationName != nil {
		runConfigurationName, err = event.RunConfigurationName.String()
		if err != nil {
			return nil, fmt.Errorf("failed to format run configuration name for proto run completion event: %w", err)
		}
	}

	var runName string
	if event.RunName != nil {
		runName, err = event.RunName.String()
		if err != nil {
			return nil, fmt.Errorf("failed to format run name for proto run completion event: %w", err)
		}
	}

	if runConfigurationName == "" && runName == "" {
		return nil, fmt.Errorf(
			"both runConfigurationName and runName are empty for the run completion event with runId: %s, pipelineName: %v",
			event.RunId,
			event.PipelineName,
		)
	}

	provider, err := event.Provider.String()
	if err != nil {
		return nil, fmt.Errorf(
			"unable to parse provider namespace name as a string for runId: %s, pipelineName: %v",
			event.RunId,
			event.PipelineName,
		)
	}

	var runStartTime *timestamppb.Timestamp
	if event.RunStartTime != nil {
		runStartTime = timestamppb.New(*event.RunStartTime)
	}

	var runEndTime *timestamppb.Timestamp
	if event.RunEndTime != nil {
		runEndTime = timestamppb.New(*event.RunEndTime)
	}

	runCompletionEvent := pb.RunCompletionEvent{
		PipelineName:          pipelineName,
		Provider:              provider,
		RunConfigurationName:  runConfigurationName,
		RunId:                 event.RunId,
		RunName:               runName,
		ServingModelArtifacts: artifactToProto(event.ServingModelArtifacts),
		Artifacts:             artifactToProto(event.Artifacts),
		Status:                statusToProto(event.Status),
		RunStartTime:          runStartTime,
		RunEndTime:            runEndTime,
	}

	return &runCompletionEvent, nil
}

func artifactToProto(commonArtifacts []common.Artifact) []*pb.Artifact {
	pbArtifacts := []*pb.Artifact{}
	for _, commonArtifact := range commonArtifacts {
		pbArtifacts = append(pbArtifacts, &pb.Artifact{
			Location: commonArtifact.Location,
			Name:     commonArtifact.Name,
		})
	}
	return pbArtifacts
}

func statusToProto(status common.RunCompletionStatus) pb.Status {
	switch status {
	case common.RunCompletionStatuses.Succeeded:
		return pb.Status_SUCCEEDED
	default:
		return pb.Status_FAILED
	}
}
