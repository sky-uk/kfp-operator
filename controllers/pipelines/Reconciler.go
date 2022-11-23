package pipelines

import (
	"context"
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
	ownerName, hasOnwerName := workflow.GetLabels()[WorkflowConstants.OwnerNameLabelKey]
	ownerNamespace, hasOnwerNamespace := workflow.GetLabels()[WorkflowConstants.OwnerNamespaceLabelKey]
	if !hasKind || !hasOnwerName || !hasOnwerNamespace || kind != resource.GetKind() {
		return nil
	}

	namespacedName := types.NamespacedName{
		Name:      ownerName,
		Namespace: ownerNamespace,
	}

	err := br.EC.Client.Cached.Get(context.TODO(), namespacedName, resource)
	if err != nil {
		return []reconcile.Request{}
	}

	return []reconcile.Request{{NamespacedName: namespacedName}}
}
