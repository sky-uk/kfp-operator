package webhook

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sky-uk/kfp-operator/argo/common"
	"google.golang.org/grpc/credentials/insecure"

	config "github.com/sky-uk/kfp-operator/apis/config/hub"
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

	runConfigurationName, err := event.RunConfigurationName.String()
	if err != nil {
		return nil, fmt.Errorf("failed to format run configuration name for proto run completion event: %w", err)
	}

	runName := ""
	if event.RunName != nil {
		runName, err = event.RunName.String()
		if err != nil {
			return nil, fmt.Errorf("failed to format run name for proto run completion event: %w", err)
		}
	}

	runCompletionEvent := pb.RunCompletionEvent{
		PipelineName:          pipelineName,
		Provider:              event.Provider,
		RunConfigurationName:  runConfigurationName,
		RunId:                 event.RunId,
		RunName:               runName,
		ServingModelArtifacts: artifactToProto(event.ServingModelArtifacts),
		Artifacts:             artifactToProto(event.Artifacts),
		Status:                statusToProto(event.Status),
		Training:              trainingToProto(event.Training),
	}

	return &runCompletionEvent, nil
}

func trainingToProto(training *common.Training) *pb.Training {
	if training == nil || training.IsEmpty() {
		return nil
	}

	pbTraining := pb.Training{}
	if training.StartTime != nil {
		pbTraining.StartTime = timestamppb.New(*training.StartTime)
	}
	if training.EndTime != nil {
		pbTraining.EndTime = timestamppb.New(*training.EndTime)
	}

	return &pbTraining

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
