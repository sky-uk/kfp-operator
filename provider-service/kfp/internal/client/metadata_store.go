package client

import (
	"context"
	"fmt"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/ml_metadata"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"strings"

	"github.com/hashicorp/go-bexpr"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1beta1"
	"github.com/sky-uk/kfp-operator/argo/common"
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
	GetArtifacts(ctx context.Context, workflowName string, artifactDefs []pipelinesv1.OutputArtifact) ([]common.Artifact, error)
}

type GrpcMetadataStore struct {
	MetadataStoreServiceClient MetadataStoreServiceClient
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

func (gms *GrpcMetadataStore) GetArtifacts(ctx context.Context, workflowName string, artifactDefs []pipelinesv1.OutputArtifact) (artifacts []common.Artifact, err error) {
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

	for _, artifactDef := range artifactDefs {
		var evaluator *bexpr.Evaluator
		if artifactDef.Path.Filter != "" {
			evaluator, err = bexpr.CreateEvaluator(artifactDef.Path.Filter)
			if err != nil {
				return nil, err
			}
		}

		for _, artifact := range artifactsResponse.GetArtifacts() {
			artifactUri := artifact.GetUri()
			if artifactUri == "" {
				continue
			}
			if !strings.HasSuffix(artifact.GetName(), artifactDef.Path.Locator.String()) {
				continue
			}

			if evaluator != nil {
				matched, err := evaluator.Evaluate(propertiesToPrimitiveMap(artifact.GetCustomProperties()))
				// evaluator errors on missing properties
				if err != nil {
					continue
				}
				if !matched {
					continue
				}
			}

			artifacts = append(artifacts, common.Artifact{
				Name:     artifactDef.Name,
				Location: artifactUri,
			})
		}
	}

	return artifacts, nil
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
		case *ml_metadata.Value_StructValue:
			out[k] = v.GetStructValue().AsMap()
		}
	}

	return out
}

func CreateMetadataStore(ctx context.Context, config config.KfpProviderConfig) (MetadataStore, error) {
	logger := common.LoggerFromContext(ctx)
	metadataStore, err := ConnectToMetadataStore(config.Parameters.GrpcMetadataStoreAddress)
	if err != nil {
		logger.Error(err, "failed to connect to metadata store", "address", config.Parameters.GrpcMetadataStoreAddress)
		return nil, err
	}
	return metadataStore, nil
}

func ConnectToMetadataStore(address string) (*GrpcMetadataStore, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GrpcMetadataStore{
		MetadataStoreServiceClient: ml_metadata.NewMetadataStoreServiceClient(conn),
	}, nil
}
