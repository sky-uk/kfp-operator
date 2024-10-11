package webhook

import (
	"context"
	"fmt"

	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	pb "github.com/sky-uk/kfp-operator/triggers/nats_event_trigger/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UpstreamService interface {
	call(ctx context.Context, ed EventData) error
}

type GrpcNatsTrigger struct {
	Upstream          config.Endpoint
	Client            pb.NATSEventTriggerClient
	ConnectionHandler func() error
}

func NewGrpcNatsTrigger(endpoint config.Endpoint) GrpcNatsTrigger {
	conn, err := grpc.NewClient(endpoint.URL(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Errorf("failed to connect to gRPC server at %s: %v", endpoint.URL(), err))
	}

	client := pb.NewNATSEventTriggerClient(conn)

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

func (gnt GrpcNatsTrigger) call(ctx context.Context, eventData EventData) error {
	runCompletionEvent, err := EventDataToPbRunCompletion(eventData)
	if err != nil {
		return err
	}
	_, err = gnt.Client.ProcessEventFeed(ctx, runCompletionEvent)
	return err
}

func EventDataToPbRunCompletion(eventData EventData) (*pb.RunCompletionEvent, error) {
	pipelineName, err := eventData.RunCompletionEvent.PipelineName.String()
	if err != nil {
		return nil, fmt.Errorf("failed to format pipeline name for proto run completion event: %w", err)
	}

	runConfigurationName, err := eventData.RunCompletionEvent.RunConfigurationName.String()
	if err != nil {
		return nil, fmt.Errorf("failed to format run configuration name for proto run completion event: %w", err)
	}

	runName, err := eventData.RunCompletionEvent.RunName.String()
	if err != nil {
		return nil, fmt.Errorf("failed to format run name for proto run completion event: %w", err)
	}

	var artifacts []*pb.ServingModelArtifact
	for _, artifact := range eventData.RunCompletionEvent.ServingModelArtifacts {
		artifacts = append(artifacts, &pb.ServingModelArtifact{
			Location: artifact.Location,
			Name:     artifact.Name,
		})
	}

	runCompletionEvent := pb.RunCompletionEvent{
		PipelineName:          pipelineName,
		Provider:              eventData.RunCompletionEvent.Provider,
		RunConfigurationName:  runConfigurationName,
		RunId:                 eventData.RunCompletionEvent.RunId,
		RunName:               runName,
		ServingModelArtifacts: artifacts,
		Status:                string(eventData.RunCompletionEvent.Status),
	}

	return &runCompletionEvent, err
}
