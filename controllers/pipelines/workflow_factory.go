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
	OwnerNamespaceLabelKey      string
	OperationLabelKey           string
	CreateOperationLabel        string
	DeleteOperationLabel        string
	UpdateOperationLabel        string
	EntryPointName              string
	ConstructionFailedError     string
	ProviderNameParameterName   string
	ProviderOutputParameterName string
}{
	OwnerKindLabelKey:           pipelinesv1.GroupVersion.Group + "/owner.kind",
	OwnerNameLabelKey:           pipelinesv1.GroupVersion.Group + "/owner.name",
	OwnerNamespaceLabelKey:      pipelinesv1.GroupVersion.Group + "/owner.namespace",
	OperationLabelKey:           pipelinesv1.GroupVersion.Group + "/operation",
	CreateOperationLabel:        "create",
	DeleteOperationLabel:        "delete",
	UpdateOperationLabel:        "update",
	EntryPointName:              "main",
	ConstructionFailedError:     "error constructing workflow",
	ProviderNameParameterName:   "provider-name",
	ProviderOutputParameterName: "provider-output",
}

type WorkflowFactory[R pipelinesv1.Resource] interface {
	ConstructCreationWorkflow(resource R) (*argo.Workflow, error)
	ConstructUpdateWorkflow(resource R) (*argo.Workflow, error)
	ConstructDeletionWorkflow(resource R) (*argo.Workflow, error)
}

type WorkflowFactoryBase struct {
	ResourceKind string
	Config       config.Configuration
}

func (w WorkflowFactoryBase) CommonWorkflowMeta(owner pipelinesv1.Resource, operation string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: fmt.Sprintf("%s-%s-", operation, owner.GetKind()),
		Namespace:    w.Config.WorkflowNamespace,
		Labels:       CommonWorkflowLabels(owner, operation),
	}
}

func CommonWorkflowLabels(owner pipelinesv1.Resource, operation string) map[string]string {
	return map[string]string{
		WorkflowConstants.OperationLabelKey:      operation,
		WorkflowConstants.OwnerKindLabelKey:      owner.GetKind(),
		WorkflowConstants.OwnerNameLabelKey:      owner.GetName(),
		WorkflowConstants.OwnerNamespaceLabelKey: owner.GetNamespace(),
	}
}
