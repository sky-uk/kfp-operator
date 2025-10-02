package client

import (
	"context"
	"fmt"
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

func (gms *GrpcMetadataStore) GetArtifactsForRun(ctx context.Context, runId string) (artifacts []common.PipelineComponent, err error) {
	typeName := "system.PipelineRun"
	ctxResp, err := gms.MetadataStoreServiceClient.GetContextByTypeAndName(ctx, &ml_metadata.GetContextByTypeAndNameRequest{
		TypeName:    &typeName,
		ContextName: &runId, // e.g., your KFP run ID
	})
	if err != nil {
		return nil, err
	}
	// 1. Fetch all executions and filter by run_id
	execResp, err := gms.MetadataStoreServiceClient.GetExecutionsByContext(ctx, &ml_metadata.GetExecutionsByContextRequest{
		ContextId: ctxResp.GetContext().Id,
	})
	if err != nil {
		return nil, err
	}
	// 2. Fetch events for those executions
	var execIds []int64
	for _, runExec := range execResp.Executions {
		execIds = append(execIds, runExec.GetId())
	}
	eventsResp, err := gms.MetadataStoreServiceClient.GetEventsByExecutionIDs(ctx, &ml_metadata.GetEventsByExecutionIDsRequest{
		ExecutionIds: execIds,
	})
	if err != nil {
		return nil, err
	}

	taskArtifactIDs := map[int64][]int64{}
	for _, ev := range eventsResp.GetEvents() {
		if ev.GetType() == ml_metadata.Event_OUTPUT {
			taskArtifactIDs[ev.GetExecutionId()] = append(taskArtifactIDs[ev.GetExecutionId()], ev.GetArtifactId())
		}
	}

	// 4. Fetch artifact details
	var allArtifactIDs []int64
	for _, ids := range taskArtifactIDs {
		allArtifactIDs = append(allArtifactIDs, ids...)
	}
	artifactsResp, err := gms.MetadataStoreServiceClient.GetArtifactsByID(ctx, &ml_metadata.GetArtifactsByIDRequest{
		ArtifactIds: allArtifactIDs,
	})
	if err != nil {
		return nil, err
	}
	// 5. Map artifact IDs -> artifact proto
	artifactByID := map[int64]*ml_metadata.Artifact{}
	for _, artifact := range artifactsResp.Artifacts {
		artifactByID[artifact.GetId()] = artifact
	}

	var pcs []common.PipelineComponent
	for _, runExec := range execResp.Executions {
		pc := common.PipelineComponent{
			Name:               runExec.CustomProperties["display_name"].GetStringValue(),
			ComponentArtifacts: []common.ComponentArtifact{},
		}

		for _, artID := range taskArtifactIDs[runExec.GetId()] {
			art := artifactByID[artID]

			pc.ComponentArtifacts = append(pc.ComponentArtifacts, common.ComponentArtifact{
				Name: art.CustomProperties["display_name"].GetStringValue(),
				Artifacts: []common.ComponentArtifactInstance{
					{
						Uri:      art.GetUri(),
						Metadata: propertiesToPrimitiveMap(art.GetProperties()),
					},
				},
			})

		}
		pcs = append(pcs, pc)
	}
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
