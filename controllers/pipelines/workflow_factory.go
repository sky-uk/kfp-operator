package pipelines

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var WorkflowConstants = struct {
	OwnerKindLabelKey           string
	OwnerNameLabelKey           string
	OperationLabelKey           string
	CreateOperationLabel        string
	DeleteOperationLabel        string
	UpdateOperationLabel        string
	EntryPointName              string
	ConstructionFailedError     string
	ProviderConfigParameterName string
	ProviderOutputParameterName string
}{
	OwnerKindLabelKey:           pipelinesv1.GroupVersion.Group + "/owner.kind",
	OwnerNameLabelKey:           pipelinesv1.GroupVersion.Group + "/owner.name",
	OperationLabelKey:           pipelinesv1.GroupVersion.Group + "/operation",
	CreateOperationLabel:        "create",
	DeleteOperationLabel:        "delete",
	UpdateOperationLabel:        "update",
	EntryPointName:              "main",
	ConstructionFailedError:     "error constructing workflow",
	ProviderConfigParameterName: "provider-config",
	ProviderOutputParameterName: "provider-output",
}

type WorkflowFactory[R pipelinesv1.Resource] interface {
	ConstructCreationWorkflow(resource R) (*argo.Workflow, error)
	ConstructUpdateWorkflow(resource R) (*argo.Workflow, error)
	ConstructDeletionWorkflow(resource R) (*argo.Workflow, error)
}

type WorkflowFactoryBase struct {
	ResourceKind   string
	Config         config.Configuration
	ProviderConfig string
}

func CommonWorkflowMeta(owner pipelinesv1.Resource, operation string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: fmt.Sprintf("%s-%s-", operation, owner.GetKind()),
		Namespace:    owner.GetNamespace(),
		Labels:       CommonWorkflowLabels(owner, operation),
	}
}

func CommonWorkflowLabels(owner pipelinesv1.Resource, operation string) map[string]string {
	return map[string]string{
		WorkflowConstants.OperationLabelKey: operation,
		WorkflowConstants.OwnerKindLabelKey: owner.GetKind(),
		WorkflowConstants.OwnerNameLabelKey: owner.GetName(),
	}
}
