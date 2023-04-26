package kfp

import (
	"context"
	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	"gopkg.in/yaml.v2"
)

var kfpApiConstants = struct {
	KfpResourceNotFoundCode int32
}{
	KfpResourceNotFoundCode: 5,
}

type KfpApi interface {
	GetResourceReferences(ctx context.Context, runId string) (ResourceReferences, error)
}

type GrpcKfpApi struct {
	RunServiceClient go_client.RunServiceClient
	JobServiceClient go_client.JobServiceClient
}

type ResourceReferences struct {
	RunConfigurationName string
	RunName              common.NamespacedName
}

func (gka *GrpcKfpApi) GetResourceReferences(ctx context.Context, runId string) (ResourceReferences, error) {
	resourceReferences := ResourceReferences{}

	runDetail, err := gka.RunServiceClient.GetRun(ctx, &go_client.GetRunRequest{RunId: runId})
	if err != nil {
		return resourceReferences, err
	}

	resourceReferences.RunName.Name = runDetail.GetRun().GetName()

	for _, ref := range runDetail.GetRun().GetResourceReferences() {
		if ref.GetKey().GetType() == go_client.ResourceType_JOB && ref.GetRelationship() == go_client.Relationship_CREATOR {
			rcNameFromJob, err := gka.GetRunConfigurationNameFromJob(ctx, ref.GetKey().GetId())
			if err != nil {
				return ResourceReferences{}, err
			}
			if rcNameFromJob == "" {
				// For migration from v1alpha4. Remove afterwards.
				resourceReferences.RunConfigurationName = ref.GetName()
			} else {
				resourceReferences.RunConfigurationName = rcNameFromJob
			}
			continue
		}

		if ref.GetKey().GetType() == go_client.ResourceType_NAMESPACE && ref.GetRelationship() == go_client.Relationship_OWNER {
			resourceReferences.RunName.Namespace = ref.GetKey().GetId()
			continue
		}
	}

	return resourceReferences, nil
}

func (gka *GrpcKfpApi) GetRunConfigurationNameFromJob(ctx context.Context, jobId string) (string, error) {
	job, err := gka.JobServiceClient.GetJob(ctx, &go_client.GetJobRequest{Id: jobId})
	if err != nil {
		return "", err
	}

	runScheduleDefinition := base.RunScheduleDefinition{}
	if err := yaml.Unmarshal([]byte(job.Description), &runScheduleDefinition); err != nil {
		return "", nil
	}

	return runScheduleDefinition.RunConfigurationName, nil
}
