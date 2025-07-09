package provider

import (
	"context"
	"fmt"

	"github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/util"
	"gopkg.in/yaml.v2"

	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

type RunService interface {
	CreateRun(
		ctx context.Context,
		rd base.RunDefinition,
		pipelineId string,
		pipelineVersionId string,
		experimentId string,
	) (string, error)
}

type DefaultRunService struct {
	client client.RunServiceClient
}

func NewRunService(
	conn *grpc.ClientConn,
) (RunService, error) {
	if conn == nil {
		return nil, fmt.Errorf(
			"no gRPC connection was provided to start run service",
		)
	}

	return &DefaultRunService{
		client: go_client.NewRunServiceClient(conn),
	}, nil
}

// CreateRun creates a run and returns the generated run id.
func (rs DefaultRunService) CreateRun(
	ctx context.Context,
	rd base.RunDefinition,
	pipelineId string,
	pipelineVersionId string,
	experimentId string,
) (string, error) {
	runParameters := make(map[string]*structpb.Value)
	for k, v := range rd.Parameters {
		runParameters[k] = structpb.NewStringValue(v)
	}

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

	run, err := rs.client.CreateRun(ctx, &go_client.CreateRunRequest{
		Run: &go_client.Run{
			ExperimentId: experimentId,
			DisplayName:  name,
			Description:  string(runAsDescription),
			PipelineSource: &go_client.Run_PipelineVersionReference{
				PipelineVersionReference: &go_client.PipelineVersionReference{
					PipelineId:        pipelineId,
					PipelineVersionId: pipelineVersionId,
				},
			},
			RuntimeConfig: &go_client.RuntimeConfig{
				Parameters: runParameters,
			},
		},
	})
	if err != nil {
		return "", err
	}

	return run.RunId, nil
}
