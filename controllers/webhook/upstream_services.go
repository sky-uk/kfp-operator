package webhook

import (
	"context"
	"fmt"
	"github.com/google/martian/log"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	pb "github.com/sky-uk/kfp-operator/triggers/nats_event_trigger/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UpstreamService interface {
	call(ctx context.Context, ed EventData) error
}

type GrpcNatsTrigger struct {
	Upstream config.Endpoint
	Client   *pb.NATSEventTriggerClient
}

func NewGrpcNatsTrigger(endpoint config.Endpoint) GrpcNatsTrigger {
	conn, err := grpc.NewClient(endpoint.URL(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Errorf("failed to connect to gRPC server at %s: %v", endpoint.URL(), err))
	}

	defer func(conn *grpc.ClientConn) {
		if err := conn.Close(); err != nil {
			log.Errorf("failed to close gRPC client connection at %s: %v", endpoint.URL(), err)
		}
	}(conn)

	client := pb.NewNATSEventTriggerClient(conn)

	return GrpcNatsTrigger{
		Upstream: endpoint,
		Client:   &client,
	}
}

func (gnt GrpcNatsTrigger) call(ctx context.Context, eventData EventData) error {
	runCompletionEvent, err := EventDataToPbRunCompletion(eventData)
	if err != nil {
		return err
	}
	_, err = pb.NATSEventTriggerClient.ProcessEventFeed(*gnt.Client, ctx, runCompletionEvent)
	return err
}

func EventDataToPbRunCompletion(eventData EventData) (*pb.RunCompletionEvent, error) {
	pipelineName, err := eventData.RunCompletionEvent.PipelineName.String()
	runConfigurationName, err := eventData.RunCompletionEvent.RunConfigurationName.String()
	runName, err := eventData.RunCompletionEvent.RunName.String()

	if err != nil {
		return nil, fmt.Errorf("failed to format event data for proto run completion event: %w", err)
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
