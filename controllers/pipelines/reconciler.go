package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
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

func desiredProvider(resource pipelinesv1.HasProvider, config config.Configuration) string {
	if provider, hasProvider := resource.GetAnnotations()[apis.ResourceAnnotations.Provider]; hasProvider {
		return provider
	}

	if provider := resource.GetProvider(); provider != "" {
		return provider
	}

	return config.DefaultProvider
}

func (br ResourceReconciler[R]) loadProvider(ctx context.Context, namespace string, desiredProvider string) (pipelinesv1.Provider, error) {
	providerNamespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      desiredProvider,
	}
	var provider = pipelinesv1.Provider{}

	err := br.EC.Client.NonCached.Get(ctx, providerNamespacedName, &provider)

	return provider, err
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
