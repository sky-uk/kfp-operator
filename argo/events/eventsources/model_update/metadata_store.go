package main

import (
	"context"
	"fmt"
	"pipelines.kubeflow.org/events/ml_metadata"
)

const (
	PushedDestinationCustomProperty = "pushed_destination"
)
type ModelArtifact struct {
	PushDestination string
}

type MetadataStore interface {
	GetServingModelArtifact(pipelineName string, workflowName string) ModelArtifact
}

type GrpcMetadataStore struct {
	MetadataStoreServiceClient ml_metadata.MetadataStoreServiceClient
}

//TODO: call GetContextByTypeAndName
//TODO: filter GetExecutionsByContext by runId(workflowName) and component('Pusher')
//TODO: filter GetEventsByExecutionIDs by stepName(pushed_model)
//TODO: GetServingModelArtifact to return collection with all suitable artifacts (not first)

func (gms *GrpcMetadataStore) GetServingModelArtifact(ctx context.Context, pipelineName string, workflowName string) ([]ModelArtifact, error) {
	executionsResponse, err := gms.MetadataStoreServiceClient.GetExecutionsByContext(ctx, &ml_metadata.GetExecutionsByContextRequest{
		ContextId: nil,
	})
	if err != nil {
		return []ModelArtifact{}, err
	}
	executions := executionsResponse.GetExecutions()
	if executions == nil {
		return []ModelArtifact{}, fmt.Errorf("missing executions")
	}

	var executionIds []int64
	for _, execution := range executions {
		executionIds = append(executionIds, execution.GetId())
	}

	eventsResponse, err := gms.MetadataStoreServiceClient.GetEventsByExecutionIDs(ctx, &ml_metadata.GetEventsByExecutionIDsRequest{
		ExecutionIds: executionIds,
	})
	if err != nil {
		return nil, err
	}
	events := eventsResponse.GetEvents()
	if events == nil {
		return nil, fmt.Errorf("missing events")
	}

	var artifactIds []int64
	for _, event := range events {
		if event.ArtifactId != nil {
			artifactIds = append(artifactIds, *event.ArtifactId)
		}
	}

	artifactsResponse, err := gms.MetadataStoreServiceClient.GetArtifactsByID(ctx, &ml_metadata.GetArtifactsByIDRequest{
		ArtifactIds: artifactIds,
	})
	if err != nil {
		return nil, err
	}

	var results []ModelArtifact
	for _, artifact := range artifactsResponse.GetArtifacts() {
		pushDestination := artifact.GetCustomProperties()[PushedDestinationCustomProperty].GetStringValue()

		if pushDestination != "" {
			results = append(results, ModelArtifact{
				PushDestination: pushDestination,
			})
		}
	}

	return results, nil
}
