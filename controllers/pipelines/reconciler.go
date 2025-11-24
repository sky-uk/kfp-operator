package pipelines

import (
	"context"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	"github.com/sky-uk/kfp-operator/internal/config"
	"github.com/sky-uk/kfp-operator/pkg/common"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ProviderLoader interface {
	LoadProvider(
		ctx context.Context,
		desiredProvider common.NamespacedName,
	) (pipelineshub.Provider, error)
}

type ResourceReconciler[R pipelineshub.Resource] struct {
	EC     K8sExecutionContext
	Config config.ConfigSpec
}

func (br ResourceReconciler[R]) LoadProvider(
	ctx context.Context,
	desiredProvider common.NamespacedName,
) (pipelineshub.Provider, error) {
	logger := log.FromContext(ctx)
	logger.V(2).Info("loading provider:", "provider", desiredProvider)

	providerNamespacedName := types.NamespacedName{
		Name:      desiredProvider.Name,
		Namespace: desiredProvider.Namespace,
	}
	var provider = pipelineshub.Provider{}

	err := br.EC.Client.NonCached.Get(ctx, providerNamespacedName, &provider)

	return provider, err
}

func (br ResourceReconciler[R]) reconciliationRequestsForWorkflow(
	resource pipelineshub.Resource,
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
