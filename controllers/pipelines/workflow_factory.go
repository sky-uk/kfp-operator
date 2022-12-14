package pipelines

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/apis"
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
	OwnerKindLabelKey:           apis.Group + "/owner.kind",
	OwnerNameLabelKey:           apis.Group + "/owner.name",
	OwnerNamespaceLabelKey:      apis.Group + "/owner.namespace",
	OperationLabelKey:           apis.Group + "/operation",
	CreateOperationLabel:        "create",
	DeleteOperationLabel:        "delete",
	UpdateOperationLabel:        "update",
	EntryPointName:              "main",
	ConstructionFailedError:     "error constructing workflow",
	ProviderNameParameterName:   "provider-name",
	ProviderOutputParameterName: "provider-output",
}

type WorkflowFactory[R pipelinesv1.Resource] interface {
	ConstructCreationWorkflow(provider string, resource R) (*argo.Workflow, error)
	ConstructUpdateWorkflow(provider string, resource R) (*argo.Workflow, error)
	ConstructDeletionWorkflow(provider string, resource R) (*argo.Workflow, error)
}

type CompiledResourceWorkflowFactory[R pipelinesv1.Resource] struct {
	WorkflowFactoryBase
	DefinitionCreator func(R) string
}

type WorkflowFactoryBase struct {
	Config config.Configuration
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
