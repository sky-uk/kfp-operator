package foo

import (
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
)

var WorkflowConstants = struct {
	OwnerKindLabelKey               string
	OwnerNameLabelKey               string
	OwnerNamespaceLabelKey          string
	ConstructionFailedError         string
	ProviderConfigParameterName     string
	ProviderNameParameterName       string
	ProviderOutputParameterName     string
	ResourceKindParameterName       string
	ResourceDefinitionParameterName string
	ResourceIdParameterName         string
}{
	OwnerKindLabelKey:               apis.Group + "/owner.kind",
	OwnerNameLabelKey:               apis.Group + "/owner.name",
	OwnerNamespaceLabelKey:          apis.Group + "/owner.namespace",
	ConstructionFailedError:         "error constructing workflow",
	ProviderConfigParameterName:     "provider-config",
	ProviderNameParameterName:       "provider-name",
	ProviderOutputParameterName:     "provider-output",
	ResourceKindParameterName:       "resource-kind",
	ResourceDefinitionParameterName: "resource-definition",
	ResourceIdParameterName:         "resource-id",
}

func CommonWorkflowLabels(owner pipelinesv1.Resource) map[string]string {
	return map[string]string{
		WorkflowConstants.OwnerKindLabelKey:      owner.GetKind(),
		WorkflowConstants.OwnerNameLabelKey:      owner.GetName(),
		WorkflowConstants.OwnerNamespaceLabelKey: owner.GetNamespace(),
	}
}
