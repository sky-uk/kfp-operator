package pipelines

import (
	"context"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ProviderLoader interface {
	LoadProvider(
		ctx context.Context,
		namespace string,
		desiredProvider string) (pipelinesv1.Provider, error)
}

type ResourceReconciler[R pipelinesv1.Resource] struct {
	EC     K8sExecutionContext
	Config config.KfpControllerConfigSpec
}

func (br ResourceReconciler[R]) LoadProvider(
	ctx context.Context,
	namespace string,
	desiredProvider string,
) (pipelinesv1.Provider, error) {
	providerNamespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      desiredProvider,
	}
	var provider = pipelinesv1.Provider{}

	err := br.EC.Client.NonCached.Get(ctx, providerNamespacedName, &provider)

	return provider, err
}

func (br ResourceReconciler[R]) reconciliationRequestsForWorkflow(
	resource pipelinesv1.Resource,
) handler.MapFunc {
	return func(ctx context.Context, workflow client.Object) []reconcile.Request {
		kind, hasKind := workflow.GetLabels()[workflowconstants.OwnerKindLabelKey]
		ownerName, hasOwnerName := workflow.GetLabels()[workflowconstants.OwnerNameLabelKey]
		ownerNamespace, hasOwnerNamespace := workflow.GetLabels()[workflowconstants.OwnerNamespaceLabelKey]
		if !hasKind || !hasOwnerName || !hasOwnerNamespace || kind != resource.GetKind() {
			return nil
		}

		return []reconcile.Request{{NamespacedName: types.NamespacedName{
			Name:      ownerName,
			Namespace: ownerNamespace,
		}}}
	}
}

func (br ResourceReconciler[R]) setupWithManager(
	controllerBuilder *builder.Builder,
	resource R,
) *builder.Builder {
	return controllerBuilder.Watches(
		&argo.Workflow{},
		handler.EnqueueRequestsFromMapFunc(br.reconciliationRequestsForWorkflow(resource)),
		builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
	)
}
