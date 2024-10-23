package webhook

import (
	"context"
	"fmt"
	"github.com/sky-uk/kfp-operator/argo/common"
	"google.golang.org/grpc/credentials/insecure"

	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	pb "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/proto"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type UpstreamService interface {
	call(ctx context.Context, ed common.RunCompletionEvent) error
}

type GrpcNatsTrigger struct {
	Upstream          config.Endpoint
	Client            pb.RunCompletionEventTriggerClient
	ConnectionHandler func() error
}

func NewGrpcNatsTrigger(ctx context.Context, endpoint config.Endpoint) GrpcNatsTrigger {
	logger := log.FromContext(ctx)

	conn, err := grpc.NewClient(endpoint.URL(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	logger.Info("NewGrpcNatsTrigger connected to", "endpoint", endpoint.URL())
	if err != nil {
		logger.Error(err, "Error creating grpc client")
	}

	client := pb.NewRunCompletionEventTriggerClient(conn)

	return GrpcNatsTrigger{
		Upstream: endpoint,
		Client:   client,
		ConnectionHandler: func() error {
			if err := conn.Close(); err != nil {
				return err
			}
			return nil
		},
	}
}

func (gnt GrpcNatsTrigger) call(ctx context.Context, event common.RunCompletionEvent) error {
	runCompletionEvent, err := RunCompletionEventToProto(event)
	if err != nil {
		return err
	}
	_, err = gnt.Client.ProcessEventFeed(ctx, runCompletionEvent)
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

	runName, err := event.RunName.String()
	if err != nil {

		return nil, fmt.Errorf("failed to format run name for proto run completion event: %w", err)
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
