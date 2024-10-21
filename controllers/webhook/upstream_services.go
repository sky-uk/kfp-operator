package webhook

import (
	"context"
	"fmt"
	"github.com/sky-uk/kfp-operator/argo/common"
	"google.golang.org/grpc/credentials/insecure"
	"log"

	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	pb "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/proto"
	"google.golang.org/grpc"
)

type UpstreamService interface {
	call(ctx context.Context, ed common.RunCompletionEvent) error
}

type GrpcNatsTrigger struct {
	Upstream          config.Endpoint
	Client            pb.RunCompletionEventTriggerClient
	ConnectionHandler func() error
}

func NewGrpcNatsTrigger(endpoint config.Endpoint) GrpcNatsTrigger {
	conn, err := grpc.NewClient(endpoint.URL(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	log.Printf("NewGrpcNatsTrigger connected to: %s", endpoint.URL())
	if err != nil {
		log.Fatal("Error creating grpc client", err)
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

	var artifacts []*pb.ServingModelArtifact
	for _, artifact := range event.ServingModelArtifacts {
		artifacts = append(artifacts, &pb.ServingModelArtifact{
			Location: artifact.Location,
			Name:     artifact.Name,
		})
	}

	runCompletionEvent := pb.RunCompletionEvent{
		PipelineName:          pipelineName,
		Provider:              event.Provider,
		RunConfigurationName:  runConfigurationName,
		RunId:                 event.RunId,
		RunName:               runName,
		ServingModelArtifacts: artifacts,
		Status:                statusToProto(event.Status),
	}

	return &runCompletionEvent, err
}

func statusToProto(status common.RunCompletionStatus) pb.Status {
	if status == common.RunCompletionStatuses.Succeeded {
		return pb.Status_SUCCEEDED
	}
	return pb.Status_FAILED
}
