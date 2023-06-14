package kfp

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-bexpr"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/argo/providers/kfp/ml_metadata"
	"strings"
)

const (
	ArtifactNameCustomProperty = "name"
	PushedCustomProperty       = "pushed"
	PipelineRunTypeName        = "pipeline_run"
	InvalidId                  = 0
)

type MetadataStore interface {
	GetServingModelArtifact(ctx context.Context, workflowName string) ([]common.Artifact, error)
	GetArtifacts(ctx context.Context, workflowName string, artifactDefs []pipelinesv1.OutputArtifact) ([]common.Artifact, error)
}

type GrpcMetadataStore struct {
	MetadataStoreServiceClient ml_metadata.MetadataStoreServiceClient
}

func (gms *GrpcMetadataStore) GetServingModelArtifact(ctx context.Context, workflowName string) ([]common.Artifact, error) {
	return gms.GetArtifacts(ctx, workflowName, []pipelinesv1.OutputArtifact{base.LegacyArtifactDefinition})
}

func (gms *GrpcMetadataStore) GetArtifacts(ctx context.Context, workflowName string, artifactDefs []pipelinesv1.OutputArtifact) ([]common.Artifact, error) {
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
	for _, artifactDef := range artifactDefs {
		for _, artifact := range artifactsResponse.GetArtifacts() {
			artifactUri := artifact.GetUri()
			artifactName := artifact.GetCustomProperties()[ArtifactNameCustomProperty].GetStringValue()
			if artifactUri == "" {
				continue
			}

			if !strings.HasSuffix(artifactName, artifactDef.Path.Locator.String()) {
				continue
			}

			if artifactDef.Path.Filter != "" {
				evaluator, err := bexpr.CreateEvaluator(artifactDef.Path.Filter)
				if err != nil {
					return nil, err
				}

				matched, err := evaluator.Evaluate(propertiesToPrimitiveMap(artifact.GetCustomProperties()))
				// evaluator errors on missing properties
				if err != nil {
					continue
				}
				if !matched {
					continue
				}
			}

			results = append(results, common.Artifact{
				Name:     artifactName,
				Location: artifactUri,
			})
		}

	}

	return results, nil
}

func propertiesToPrimitiveMap(in map[string]*ml_metadata.Value) map[string]interface{} {
	out := map[string]interface{}{}

	for k, v := range in {
		switch interface{}(v.GetValue()).(type) {
		case *ml_metadata.Value_IntValue:
			out[k] = v.GetIntValue()
		case *ml_metadata.Value_StringValue:
			out[k] = v.GetStringValue()
		case *ml_metadata.Value_DoubleValue:
			out[k] = v.GetDoubleValue()
			//case 	ml_metadata.Value_StructValue:
			//	out[k] = v.GetStructValue()
		}
	}

	return out
}
