package webhook

import (
	"context"
	"fmt"
	"github.com/sky-uk/kfp-operator/argo/common"
	"google.golang.org/grpc/credentials/insecure"

	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pb "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/proto"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type RunCompletionEventTrigger struct {
	EndPoint          config.Endpoint
	Client            pb.RunCompletionEventTriggerClient
	ConnectionHandler func() error
}

func NewRuntimeCompletionEventTrigger(ctx context.Context, endpoint config.Endpoint) RunCompletionEventTrigger {
	logger := log.FromContext(ctx)

	conn, err := grpc.NewClient(endpoint.URL(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	logger.Info("RunCompletionEventTrigger client connected to", "endpoint", endpoint.URL())
	if err != nil {
		logger.Error(err, "Error creating grpc client")
	}

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

func (rcet RunCompletionEventTrigger) handle(ctx context.Context, event common.RunCompletionEvent) error {
	runCompletionEvent, err := RunCompletionEventToProto(event)
	if err != nil {
		return err
	}
	_, err = rcet.Client.ProcessEventFeed(ctx, runCompletionEvent)
	return err
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
	}

	return &runCompletionEvent, err
}

func artifactToProto(commonArtifacts []common.Artifact) []*pb.Artifact {
	var pbArtifacts []*pb.Artifact
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
