package main

import (
	"context"
	"fmt"
	"pipelines.kubeflow.org/events/ml_metadata"
)

const (
	PushedDestinationCustomProperty = "pushed_destination"
	NameCustomProperty = "name"
	PipelineRunTypeName = "pipeline_run"
	InvalidContextId = 0
)

type ModelArtifact struct {
	Name string
	PushDestination string
}

type MetadataStore interface {
	GetServingModelArtifact(pipelineName string, workflowName string) ModelArtifact
}

type GrpcMetadataStore struct {
	MetadataStoreServiceClient ml_metadata.MetadataStoreServiceClient
}

func (gms *GrpcMetadataStore) GetServingModelArtifact(ctx context.Context, workflowName string) ([]ModelArtifact, error) {
	pipelineRunTypeName := PipelineRunTypeName
	contextResponse, err := gms.MetadataStoreServiceClient.GetContextByTypeAndName(ctx, &ml_metadata.GetContextByTypeAndNameRequest{TypeName: &pipelineRunTypeName, ContextName: &workflowName})
	if err != nil {
		return nil, err
	}

	contextId := contextResponse.GetContext().GetId()
	if contextId == InvalidContextId {
		return nil, fmt.Errorf("invalid context ID")
	}

	artifactsResponse, err := gms.MetadataStoreServiceClient.GetArtifactsByContext(ctx, &ml_metadata.GetArtifactsByContextRequest{
		ContextId: &contextId,
	})
	if err != nil {
		return nil, err
	}

	var results []ModelArtifact
	for _, artifact := range artifactsResponse.GetArtifacts() {
		propertyValue := artifact.GetCustomProperties()[PushedDestinationCustomProperty].GetStringValue()
		propertyName := artifact.GetCustomProperties()[NameCustomProperty].GetStringValue()

		if propertyName != "" && propertyValue != "" {
			results = append(results, ModelArtifact{
				Name: propertyName,
				PushDestination: propertyValue,
			})
		}
	}

	return results, nil
}
