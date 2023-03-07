package kfp

import (
	"context"
	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/sky-uk/kfp-operator/argo/common"
)

var kfpApiConstants = struct {
	KfpResourceNotFoundCode     int32
}{
	KfpResourceNotFoundCode:     5,
}


type KfpApi interface {
	GetResourceReferences(ctx context.Context, runId string) (ResourceReferences, error)
}

type GrpcKfpApi struct {
	RunServiceClient go_client.RunServiceClient
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
			resourceReferences.RunConfigurationName = ref.GetName()
			continue
		}

		if ref.GetKey().GetType() == go_client.ResourceType_NAMESPACE && ref.GetRelationship() == go_client.Relationship_OWNER {
			resourceReferences.RunName.Namespace = ref.GetKey().GetId()
			continue
		}
	}

	return resourceReferences, nil
}
