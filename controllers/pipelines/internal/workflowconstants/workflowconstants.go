package workflowconstants

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

const (
	OwnerKindLabelKey                   = apis.Group + "/owner.kind"
	OwnerNameLabelKey                   = apis.Group + "/owner.name"
	OwnerNamespaceLabelKey              = apis.Group + "/owner.namespace"
	ConstructionFailedError             = "error constructing workflow"
	ProviderNameParameterName           = "provider-name"
	ProviderServiceUrl                  = "provider-service-url"
	ProviderOutputParameterName         = "provider-output"
	ResourceKindParameterName           = "resource-kind"
	ResourceDefinitionParameterName     = "resource-definition"
	ResourceIdParameterName             = "resource-id"
	PipelineFrameworkImageParameterName = "pipeline-framework-image"
	TriggeredByIndicatorSource          = "triggered-by-indicator-source"
	TriggeredByIndicatorType            = "triggered-by-indicator-type"
	TriggeredByIndicatorSourceNamespace = "triggered-by-indicator-namespace"
)

type WorkflowParameterError struct {
	SubError string
}

func (we *WorkflowParameterError) Error() string {
	return fmt.Sprintf("error in workflow: %s", we.SubError)
}

func CommonWorkflowLabels(owner pipelineshub.Resource) map[string]string {
	return map[string]string{
		OwnerKindLabelKey:      owner.GetKind(),
		OwnerNameLabelKey:      owner.GetName(),
		OwnerNamespaceLabelKey: owner.GetNamespace(),
	}
}
