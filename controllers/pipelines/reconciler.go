package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type ResourceReconciler[R pipelinesv1.Resource] struct {
	EC     K8sExecutionContext
	Config config.Configuration
}

func (br ResourceReconciler[R]) desiredProvider(resource R) string {
	if provider, hasProvider := resource.GetAnnotations()[apis.ResourceAnnotations.Provider]; hasProvider {
		return provider
	}

	return br.Config.DefaultProvider
}

func (br ResourceReconciler[R]) reconciliationRequestsForWorkflow(resource pipelinesv1.Resource) func(client.Object) []reconcile.Request {
	return func(workflow client.Object) []reconcile.Request {
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
}

func (br ResourceReconciler[R]) setupWithManager(controllerBuilder *builder.Builder, resource R) *builder.Builder {
	return controllerBuilder.Watches(&source.Kind{Type: &argo.Workflow{}},
		handler.EnqueueRequestsFromMapFunc(br.reconciliationRequestsForWorkflow(resource)),
		builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
	)
}
