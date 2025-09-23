package provider

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/label"

	"github.com/sky-uk/kfp-operator/pkg/providers/base"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/util"

	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
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
	client         client.RunServiceClient
	labelGenerator label.LabelGen
}

func NewRunService(
	conn *grpc.ClientConn,
	labelGenerator label.LabelGen,
) (RunService, error) {
	if conn == nil {
		return nil, fmt.Errorf(
			"no gRPC connection was provided to start run service",
		)
	}

	return &DefaultRunService{
		client:         go_client.NewRunServiceClient(conn),
		labelGenerator: labelGenerator,
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
	generatedLabels, err := rs.labelGenerator.GenerateLabels(rd)
	if err != nil {
		return "", err
	}

	runParameters := make(map[string]*structpb.Value)
	for k, v := range lo.Assign(rd.Parameters, generatedLabels) {
		runParameters[k] = structpb.NewStringValue(v)
	}

	name, err := util.ResourceNameFromNamespacedName(rd.Name)
	if err != nil {
		return "", err
	}

	run, err := rs.client.CreateRun(ctx, &go_client.CreateRunRequest{
		Run: &go_client.Run{
			ExperimentId: experimentId,
			DisplayName:  name,
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
