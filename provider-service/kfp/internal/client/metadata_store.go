package client

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"github.com/sky-uk/kfp-operator/internal/log"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/ml_metadata"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/sky-uk/kfp-operator/pkg/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	PushedModelArtifactType    = "PushedModel"
	ArtifactNameCustomProperty = "name"
	PushedCustomProperty       = "pushed"
	PipelineRunTypeName        = "system.PipelineRun"
	InvalidId                  = 0
	DisplayName                = "display_name"
)

type MetadataStore interface {
	GetServingModelArtifact(ctx context.Context, workflowName string) ([]common.Artifact, error)
	GetArtifactsForRun(ctx context.Context, runId string) ([]common.PipelineComponent, error)
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

func (gms *GrpcMetadataStore) GetArtifactsForRun(ctx context.Context, runId string) ([]common.PipelineComponent, error) {
	// Resolve run context
	typeName := PipelineRunTypeName
	ctxResp, err := gms.MetadataStoreServiceClient.GetContextByTypeAndName(ctx,
		&ml_metadata.GetContextByTypeAndNameRequest{TypeName: &typeName, ContextName: &runId})
	if err != nil {
		return nil, err
	}

	contextForRun := ctxResp.GetContext()
	if contextForRun == nil {
		return nil, fmt.Errorf("context not found for runId: %s", runId)
	}

	// Fetch executions for the context
	execResp, err := gms.MetadataStoreServiceClient.GetExecutionsByContext(ctx,
		&ml_metadata.GetExecutionsByContextRequest{ContextId: contextForRun.Id})
	if err != nil {
		return nil, err
	}
	executions := execResp.Executions
	execIDs := lo.Map(executions, func(e *ml_metadata.Execution, _ int) int64 { return e.GetId() })

	// Fetch events and their exported artifactIds for those executions
	eventsResp, err := gms.MetadataStoreServiceClient.GetEventsByExecutionIDs(ctx,
		&ml_metadata.GetEventsByExecutionIDsRequest{ExecutionIds: execIDs})
	if err != nil {
		return nil, err
	}
	taskArtifactIDs := lo.GroupBy(
		lo.Filter(eventsResp.GetEvents(), func(ev *ml_metadata.Event, _ int) bool {
			return ev.GetType() == ml_metadata.Event_OUTPUT
		}),
		func(ev *ml_metadata.Event) int64 { return ev.GetExecutionId() },
	)

	// Flatten artifact IDs
	var allArtifactIDs []int64
	for _, evs := range taskArtifactIDs {
		allArtifactIDs = append(allArtifactIDs,
			lo.Map(evs, func(ev *ml_metadata.Event, _ int) int64 { return ev.GetArtifactId() })...,
		)
	}

	// Fetch artifacts by the ids
	artsResp, err := gms.MetadataStoreServiceClient.GetArtifactsByID(ctx,
		&ml_metadata.GetArtifactsByIDRequest{ArtifactIds: allArtifactIDs})
	if err != nil {
		return nil, err
	}
	artifactByID := lo.Associate(artsResp.Artifacts, func(a *ml_metadata.Artifact) (int64, *ml_metadata.Artifact) {
		return a.GetId(), a
	})

	// Build PipelineComponents from executions and their artifacts
	pcs := lo.FilterMap(executions, func(exec *ml_metadata.Execution, _ int) (common.PipelineComponent, bool) {
		if exec.CustomProperties == nil || exec.CustomProperties[DisplayName] == nil {
			return common.PipelineComponent{}, false
		}
		pc := common.PipelineComponent{
			Name:               exec.CustomProperties[DisplayName].GetStringValue(),
			ComponentArtifacts: []common.ComponentArtifact{},
		}

		// Convert artifact IDs per execution into artifacts
		pc.ComponentArtifacts = lo.FilterMap(taskArtifactIDs[exec.GetId()],
			func(ev *ml_metadata.Event, _ int) (common.ComponentArtifact, bool) {
				if ev.ArtifactId == nil {
					return common.ComponentArtifact{}, false
				}
				art := artifactByID[*ev.ArtifactId]
				if art == nil {
					return common.ComponentArtifact{}, false
				}
				if art.CustomProperties == nil || art.CustomProperties[DisplayName] == nil {
					return common.ComponentArtifact{}, false
				}
				return common.ComponentArtifact{
					Name: art.CustomProperties[DisplayName].GetStringValue(),
					Artifacts: []common.ComponentArtifactInstance{{
						Uri:      art.GetUri(),
						Metadata: propertiesToPrimitiveMap(art.GetProperties()),
					}},
				}, true
			})

		return pc, len(pc.ComponentArtifacts) > 0
	})

	return pcs, nil
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

func CreateMetadataStore(ctx context.Context, config config.Config) (MetadataStore, error) {
	logger := log.LoggerFromContext(ctx)
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
