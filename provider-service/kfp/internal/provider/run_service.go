package provider

import (
	"context"
	"fmt"
	"sort"

	baseResource "github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/util"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"gopkg.in/yaml.v2"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	"google.golang.org/grpc"
)

type RunService interface {
	CreateRun(
		rd baseResource.RunDefinition,
		pipelineId string,
		pipelineVersionId string,
		experimentId string,
	) (string, error)
}

type DefaultRunService struct {
	ctx    context.Context
	client client.RunServiceClient
}

func NewRunService(
	ctx context.Context,
	conn *grpc.ClientConn,
) (RunService, error) {
	if conn == nil {
		return nil, fmt.Errorf(
			"no gRPC connection was provided to start run service",
		)
	}

	return &DefaultRunService{
		ctx:    ctx,
		client: go_client.NewRunServiceClient(conn),
	}, nil
}

// CreateRun creates a run and returns the generated run id.
func (drs DefaultRunService) CreateRun(
	rd baseResource.RunDefinition,
	pipelineId string,
	pipelineVersionId string,
	experimentId string,
) (string, error) {
	jobParameters := make([]*go_client.Parameter, 0, len(rd.RuntimeParameters))
	for name, value := range rd.RuntimeParameters {
		jobParameters = append(jobParameters, &go_client.Parameter{Name: name, Value: value})
	}
	// Sort the parameters by name for consistent ordering
	// Solely for making testing easier.
	sort.Slice(jobParameters, func(i, j int) bool {
		return jobParameters[i].Name < jobParameters[j].Name
	})

	runAsDescription, err := yaml.Marshal(resource.References{
		RunName:              rd.Name,
		RunConfigurationName: rd.RunConfigurationName,
		PipelineName:         rd.PipelineName,
		Artifacts:            rd.Artifacts,
	})
	if err != nil {
		return "", err
	}

	name, err := util.ResourceNameFromNamespacedName(rd.Name)
	if err != nil {
		return "", err
	}

	run, err := drs.client.CreateRun(drs.ctx, &go_client.CreateRunRequest{
		Run: &go_client.Run{
			Name:        name,
			Description: string(runAsDescription),
			PipelineSpec: &go_client.PipelineSpec{
				PipelineId: pipelineId,
				Parameters: jobParameters,
			},
			ResourceReferences: []*go_client.ResourceReference{
				{
					Key: &go_client.ResourceKey{
						Type: go_client.ResourceType_EXPERIMENT,
						Id:   experimentId,
					},
					Relationship: go_client.Relationship_OWNER,
				},
				{
					Key: &go_client.ResourceKey{
						Type: go_client.ResourceType_PIPELINE_VERSION,
						Id:   pipelineVersionId,
					},
					Relationship: go_client.Relationship_CREATOR,
				},
				{
					Key: &go_client.ResourceKey{
						Type: go_client.ResourceType_NAMESPACE,
						Id:   rd.Name.Namespace,
					},
					Relationship: go_client.Relationship_OWNER,
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	return run.Run.Id, nil
}
