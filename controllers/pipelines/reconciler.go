package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type BaseReconciler[R pipelinesv1.Resource] struct {
	EC           K8sExecutionContext
	StateHandler StateHandler[R]
	Config       config.Configuration
}

func (br BaseReconciler[R]) desiredProvider(resource R) string {
	if provider, hasProvider := resource.GetAnnotations()[apis.ResourceAnnotations.Provider]; hasProvider {
		return provider
	}

	return br.Config.DefaultProvider
}

func (br BaseReconciler[R]) reconciliationRequestsForWorkflow(resource R) func(object client.Object) []reconcile.Request {
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

func (br BaseReconciler[R]) setupWithManager(mgr ctrl.Manager, resource R) (*builder.Builder, error) {
	return ctrl.NewControllerManagedBy(mgr).
		For(resource).
		Watches(&source.Kind{Type: &argo.Workflow{}},
			handler.EnqueueRequestsFromMapFunc(br.reconciliationRequestsForWorkflow(resource)),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		), nil
}
