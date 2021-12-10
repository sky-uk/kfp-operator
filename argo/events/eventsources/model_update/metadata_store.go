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

func (gms *GrpcMetadataStore) GetServingModelArtifact(ctx context.Context, pipelineName string, workflowName string) (ModelArtifact, error) {
	artifactsResponse, err := gms.MetadataStoreServiceClient.GetArtifactsByID(ctx, &ml_metadata.GetArtifactsByIDRequest{
		ArtifactIds: nil,
	})

	if err != nil {
		return ModelArtifact{}, err
	}

	for _, artifact := range artifactsResponse.GetArtifacts() {
		pushDestination := artifact.GetCustomProperties()[PushedDestinationCustomProperty].GetStringValue()

		if pushDestination != "" {
			return ModelArtifact{
				PushDestination: pushDestination,
			}, nil
		}
	}

	return ModelArtifact{}, fmt.Errorf("missing artifact with push destination in response")
}
