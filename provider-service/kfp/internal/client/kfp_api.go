package client

import (
	"context"
	"github.com/sky-uk/kfp-operator/internal/log"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/label"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
)

type KfpApi interface {
	GetResourceReferences(ctx context.Context, runId string) (resource.References, error)
}

type GrpcKfpApi struct {
	RunServiceClient
	RecurringRunServiceClient
}

func (gka *GrpcKfpApi) GetResourceReferences(ctx context.Context, runId string) (resource.References, error) {
	resourceReferences := resource.References{}

	run, err := gka.RunServiceClient.GetRun(ctx, &go_client.GetRunRequest{RunId: runId})
	if err != nil {
		return resource.References{}, err
	}

	params := run.RuntimeConfig.GetParameters()
	if params != nil {
		if runName, ok := params[label.RunName]; ok {
			resourceReferences.RunName.Name = runName.GetStringValue()
		}

		if runNamespace, ok := params[label.RunNamespace]; ok {
			resourceReferences.RunName.Namespace = runNamespace.GetStringValue()
		}

		if runConfigurationName, ok := params[label.RunConfigurationName]; ok {
			resourceReferences.RunConfigurationName.Name = runConfigurationName.GetStringValue()
		}

		if runConfigurationNamespace, ok := params[label.RunConfigurationNamespace]; ok {
			resourceReferences.RunConfigurationName.Namespace = runConfigurationNamespace.GetStringValue()
		}

		if pipelineName, ok := params[label.PipelineName]; ok {
			resourceReferences.PipelineName.Name = pipelineName.GetStringValue()
		}

		if pipelineNamespace, ok := params[label.PipelineNamespace]; ok {
			resourceReferences.PipelineName.Name = pipelineNamespace.GetStringValue()
		}
	}

	if run.CreatedAt != nil {
		runCreateTime := run.CreatedAt.AsTime()
		resourceReferences.CreatedAt = &runCreateTime
	}

	if run.FinishedAt != nil {
		runFinishedTime := run.FinishedAt.AsTime()
		resourceReferences.FinishedAt = &runFinishedTime
	}

	outputArtifactIdByTaskName := map[string]map[string][]int64{}
	for _, runTask := range run.RunDetails.GetTaskDetails() {
		runTask.GetExecutorDetail().GetMainJob()

		log.LoggerFromContext(ctx).Info("found task", "task", runTask.DisplayName, "output", runTask.GetOutputs())
		for artifactName, artifacts := range runTask.Outputs {
			outputArtifactIdByTaskName[runTask.GetDisplayName()] = map[string][]int64{
				artifactName: artifacts.GetArtifactIds(),
			}
		}
	}

	resourceReferences.Artifacts = outputArtifactIdByTaskName

	return resourceReferences, nil
}

func CreateKfpApi(ctx context.Context, config config.Config) (KfpApi, error) {
	logger := log.LoggerFromContext(ctx)
	kfpApi, err := ConnectToKfpApi(config.Parameters.GrpcKfpApiAddress)
	if err != nil {
		logger.Error(err, "failed to connect to Kubeflow API", "address", config.Parameters.GrpcKfpApiAddress)
		return nil, err
	}
	return kfpApi, nil
}

func ConnectToKfpApi(address string) (*GrpcKfpApi, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GrpcKfpApi{
		RunServiceClient:          go_client.NewRunServiceClient(conn),
		RecurringRunServiceClient: go_client.NewRecurringRunServiceClient(conn),
	}, nil
}
