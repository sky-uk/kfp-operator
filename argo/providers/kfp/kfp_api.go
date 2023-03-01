package kfp

import (
	"context"
	"fmt"
	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/sky-uk/kfp-operator/argo/common"
	"strings"
)

var kfpApiConstants = struct {
	RunNameScheme               string
	ResourceIdentifierDelimiter string
	KfpResourceNotFoundCode     int32
}{
	RunNameScheme:               "pipelines.kubeflow.org/run-name",
	ResourceIdentifierDelimiter: ":",
	KfpResourceNotFoundCode:     5,
}

type ResourceIdentifier struct {
	Scheme string
	Path   string
}

func ResourceIdentifierFromString(input string) (ResourceIdentifier, error) {
	splits := strings.Split(input, kfpApiConstants.ResourceIdentifierDelimiter)

	if len(splits) < 2 || splits[0] == "" || splits[1] == "" {
		return ResourceIdentifier{}, fmt.Errorf("identifier must be in the format 'scheme:path'")
	}

	return ResourceIdentifier{
		Scheme: splits[0],
		Path:   strings.Join(splits[1:], kfpApiConstants.ResourceIdentifierDelimiter),
	}, nil
}

func (ri ResourceIdentifier) String() string {
	return fmt.Sprintf("%s:%s", ri.Scheme, ri.Path)
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
