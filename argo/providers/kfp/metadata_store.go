package kfp

import (
	"context"
	"fmt"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/argo/providers/kfp/ml_metadata"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	PushedModelArtifactType    = "PushedModel"
	ArtifactNameCustomProperty = "name"
	PushedCustomProperty       = "pushed"
	PipelineRunTypeName        = "pipeline_run"
	InvalidId                  = 0
)

type MetadataStore interface {
	GetServingModelArtifact(ctx context.Context, workflowName string) ([]common.Artifact, error)
}

type GrpcMetadataStore struct {
	MetadataStoreServiceClient ml_metadata.MetadataStoreServiceClient
}

func (gms *GrpcMetadataStore) GetServingModelArtifact(ctx context.Context, workflowName string) ([]common.Artifact, error) {
	artifactTypeName := PushedModelArtifactType
	typeResponse, err := gms.MetadataStoreServiceClient.GetArtifactType(ctx, &ml_metadata.GetArtifactTypeRequest{TypeName: &artifactTypeName})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}

		return nil, err
	}

	artifactTypeId := typeResponse.GetArtifactType().GetId()
	if artifactTypeId == InvalidId {
		return nil, fmt.Errorf("invalid artifact ID")
	}

	pipelineRunTypeName := PipelineRunTypeName
	contextResponse, err := gms.MetadataStoreServiceClient.GetContextByTypeAndName(ctx, &ml_metadata.GetContextByTypeAndNameRequest{TypeName: &pipelineRunTypeName, ContextName: &workflowName})
	if err != nil {
		return nil, err
	}
	contextId := contextResponse.GetContext().GetId()
	if contextId == InvalidId {
		return nil, fmt.Errorf("invalid context ID")
	}

	artifactsResponse, err := gms.MetadataStoreServiceClient.GetArtifactsByContext(ctx, &ml_metadata.GetArtifactsByContextRequest{
		ContextId: &contextId,
	})
	if err != nil {
		return nil, err
	}

	results := make([]common.Artifact, 0)
	for _, artifact := range artifactsResponse.GetArtifacts() {
		if artifact.GetTypeId() == artifactTypeId {
			artifactUri := artifact.GetUri()
			artifactName := artifact.GetCustomProperties()[ArtifactNameCustomProperty].GetStringValue()
			modelHasBeenPushed := artifact.GetCustomProperties()[PushedCustomProperty].GetIntValue()

			if artifactName != "" && artifactUri != "" && modelHasBeenPushed == 1 {
				results = append(results, common.Artifact{
					Name:     artifactName,
					Location: artifactUri,
				})
			}
		}
	}

	return results, nil
}
