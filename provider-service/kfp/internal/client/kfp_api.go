package client

import (
	"context"
	resource "github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/sky-uk/kfp-operator/argo/common"
	"gopkg.in/yaml.v2"
)

type KfpApi interface {
	GetResourceReferences(ctx context.Context, runId string) (resource.References, error)
}

type GrpcKfpApi struct {
	RunServiceClient
	JobServiceClient
}

func (gka *GrpcKfpApi) GetResourceReferences(ctx context.Context, runId string) (resource.References, error) {
	resourceReferences := resource.References{}

	runDetail, err := gka.RunServiceClient.GetRun(ctx, &go_client.GetRunRequest{RunId: runId})
	if err != nil {
		return resource.References{}, err
	}
	resourceReferences, ok, err := gka.GetResourceReferencesFromDescription(runDetail.Run.Description)
	if ok || err != nil {
		return resourceReferences, err
	}

	for _, ref := range runDetail.GetRun().GetResourceReferences() {
		if ref.GetKey().GetType() == go_client.ResourceType_JOB && ref.GetRelationship() == go_client.Relationship_CREATOR {
			job, err := gka.JobServiceClient.GetJob(ctx, &go_client.GetJobRequest{Id: ref.GetKey().GetId()})
			if err != nil {
				return resource.References{}, err
			}
			resourceReferences, ok, err = gka.GetResourceReferencesFromDescription(job.Description)
			if ok || err != nil {
				return resourceReferences, err
			}

			// For compatability with resources created with v0.3.0 and older
			// Pipeline name set by caller
			resourceReferences.RunConfigurationName.Name = ref.GetName()
			continue
		}

		if ref.GetKey().GetType() == go_client.ResourceType_NAMESPACE && ref.GetRelationship() == go_client.Relationship_OWNER {
			// For compatability with resources created with v0.3.0 and older
			// Pipeline name set by caller
			resourceReferences.RunName.Name = runDetail.GetRun().GetName()
			resourceReferences.RunName.Namespace = ref.GetKey().GetId()
			continue
		}
	}

	if run := runDetail.GetRun(); run != nil {
		if run.CreatedAt != nil {
			t := run.CreatedAt.AsTime()
			resourceReferences.CreatedAt = &t
		}
		if run.FinishedAt != nil {
			t := run.FinishedAt.AsTime()
			resourceReferences.FinishedAt = &t
		}
	}

	return resourceReferences, nil
}

func (gka *GrpcKfpApi) GetResourceReferencesFromDescription(description string) (resource.References, bool, error) {
	if description == "" {
		return resource.References{}, false, nil
	}

	resourceReferences := resource.References{}
	err := yaml.Unmarshal([]byte(description), &resourceReferences)
	return resourceReferences, true, err
}

func CreateKfpApi(ctx context.Context, config config.KfpProviderConfig) (KfpApi, error) {
	logger := common.LoggerFromContext(ctx)
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
		RunServiceClient: go_client.NewRunServiceClient(conn),
		JobServiceClient: go_client.NewJobServiceClient(conn),
	}, nil
}
