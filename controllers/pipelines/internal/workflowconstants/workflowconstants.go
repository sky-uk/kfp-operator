package workflowconstants

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
)

const (
	OwnerKindLabelKey                   = apis.Group + "/owner.kind"
	OwnerNameLabelKey                   = apis.Group + "/owner.name"
	OwnerNamespaceLabelKey              = apis.Group + "/owner.namespace"
	ConstructionFailedError             = "error constructing workflow"
	ProviderConfigParameterName         = "provider-config"
	ProviderNameParameterName           = "provider-name"
	ProviderServiceUrl                  = "provider-service-url"
	ProviderOutputParameterName         = "provider-output"
	ResourceKindParameterName           = "resource-kind"
	ResourceDefinitionParameterName     = "resource-definition"
	ResourceIdParameterName             = "resource-id"
	PipelineFrameworkImageParameterName = "pipeline-framework-image"
)

type WorkflowParameterError struct {
	SubError string
}

func (we *WorkflowParameterError) Error() string {
	return fmt.Sprintf("error in workflow: %s", we.SubError)
}

func CommonWorkflowLabels(owner pipelinesv1.Resource) map[string]string {
	return map[string]string{
		OwnerKindLabelKey:      owner.GetKind(),
		OwnerNameLabelKey:      owner.GetName(),
		OwnerNamespaceLabelKey: owner.GetNamespace(),
	}
}
