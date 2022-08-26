package pipelines

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/apis"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1alpha3"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var WorkflowConstants = struct {
	OwnerKindLabelKey        string
	OwnerNameLabelKey        string
	OperationLabelKey        string
	CreateOperationLabel     string
	DeleteOperationLabel     string
	UpdateOperationLabel     string
	EntryPointName           string
	ConstructionFailedError  string
	KfpEndpointParameterName string
}{
	OwnerKindLabelKey:        pipelinesv1.GroupVersion.Group + "/owner.kind",
	OwnerNameLabelKey:        pipelinesv1.GroupVersion.Group + "/owner.name",
	OperationLabelKey:        pipelinesv1.GroupVersion.Group + "/operation",
	CreateOperationLabel:     "create",
	DeleteOperationLabel:     "delete",
	UpdateOperationLabel:     "update",
	EntryPointName:           "main",
	ConstructionFailedError:  "error constructing workflow",
	KfpEndpointParameterName: "kfp-endpoint",
}

type WorkflowFactory[R apis.Resource] interface {
	ConstructCreationWorkflow(resource R) (*argo.Workflow, error)
	ConstructUpdateWorkflow(resource R) (*argo.Workflow, error)
	ConstructDeletionWorkflow(resource R) (*argo.Workflow, error)
}

type WorkflowFactoryBase struct {
	ResourceKind string
	Config       configv1.Configuration
}

type KfpExtCommandBuilder struct {
	commandParts []string
	error        error
}

func CommonWorkflowMeta(owner apis.Resource, operation string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: fmt.Sprintf("%s-%s-", operation, owner.GetKind()),
		Namespace:    owner.GetNamespace(),
		Labels:       CommonWorkflowLabels(owner, operation),
	}
}

func CommonWorkflowLabels(owner apis.Resource, operation string) map[string]string {
	return map[string]string{
		WorkflowConstants.OperationLabelKey: operation,
		WorkflowConstants.OwnerKindLabelKey: owner.GetKind(),
		WorkflowConstants.OwnerNameLabelKey: owner.GetName(),
	}
}
