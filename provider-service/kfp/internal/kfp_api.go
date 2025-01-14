package internal

import (
	"context"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	"gopkg.in/yaml.v2"
)

type KfpApi interface {
	GetResourceReferences(ctx context.Context, runId string) (ResourceReferences, error)
}

type GrpcKfpApi struct {
	RunServiceClient client.RunServiceClient
	JobServiceClient client.JobServiceClient
}

type ResourceReferences struct {
	PipelineName         common.NamespacedName        `yaml:"pipelineName"`
	RunConfigurationName common.NamespacedName        `yaml:"runConfigurationName"`
	RunName              common.NamespacedName        `yaml:"runName"`
	Artifacts            []pipelinesv1.OutputArtifact `yaml:"artifacts,omitempty"`
}

func (gka *GrpcKfpApi) GetResourceReferences(ctx context.Context, runId string) (ResourceReferences, error) {
	resourceReferences := ResourceReferences{}

	runDetail, err := gka.RunServiceClient.GetRun(ctx, &go_client.GetRunRequest{RunId: runId})
	if err != nil {
		return ResourceReferences{}, err
	}
	resourceReferences, ok, err := gka.GetResourceReferencesFromDescription(runDetail.Run.Description)
	if ok || err != nil {
		return resourceReferences, err
	}

	for _, ref := range runDetail.GetRun().GetResourceReferences() {
		if ref.GetKey().GetType() == go_client.ResourceType_JOB && ref.GetRelationship() == go_client.Relationship_CREATOR {
			job, err := gka.JobServiceClient.GetJob(ctx, &go_client.GetJobRequest{Id: ref.GetKey().GetId()})
			if err != nil {
				return ResourceReferences{}, err
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

	return resourceReferences, nil
}

func (gka *GrpcKfpApi) GetResourceReferencesFromDescription(description string) (ResourceReferences, bool, error) {
	if description == "" {
		return ResourceReferences{}, false, nil
	}

	resourceReferences := ResourceReferences{}
	err := yaml.Unmarshal([]byte(description), &resourceReferences)
	return resourceReferences, true, err
}
