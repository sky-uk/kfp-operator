package pipelines

import (
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type BaseReconciler[R pipelinesv1.Resource] struct {
	EC           K8sExecutionContext
	StateHandler StateHandler[R]
}

func (br BaseReconciler[R]) reconciliationRequestsWorkflow(workflow client.Object, resource R) []reconcile.Request {
	kind, hasKind := workflow.GetLabels()[WorkflowConstants.OwnerKindLabelKey]
	ownerName, hasOwnerName := workflow.GetLabels()[WorkflowConstants.OwnerNameLabelKey]
	ownerNamespace, hasOwnerNamespace := workflow.GetLabels()[WorkflowConstants.OwnerNamespaceLabelKey]
	if !hasKind || !hasOwnerName || !hasOwnerNamespace || kind != resource.GetKind() {
		return nil
	}

	return []reconcile.Request{{NamespacedName: types.NamespacedName{
		Name:      ownerName,
		Namespace: ownerNamespace,
	}}}
}
