package pipelines

import (
	"context"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	rcRefField = ".spec.runtimeParameters.valueFrom.runConfigurationRef.name"
)

type DependingOnRunConfigurationResource interface {
	client.Object
	GetReferencedDependencies() []string
	GetDependencyRun(string) (pipelinesv1.RunReference, bool)
	SetDependencyRun(string, pipelinesv1.RunReference)
}

type DependingOnRunConfigurationReconciler[R DependingOnRunConfigurationResource] struct {
	EC K8sExecutionContext
}

func (dr DependingOnRunConfigurationReconciler[R]) handleDependentRuns(ctx context.Context, resource R) (bool, error) {
	logger := log.FromContext(ctx)

	for _, rcName := range resource.GetReferencedDependencies() {
		runConfiguration, err := dr.getIgnoreNotFound(ctx, types.NamespacedName{
			Namespace: resource.GetNamespace(),
			Name:      rcName,
		})
		if err != nil || runConfiguration == nil {
			return false, err
		}

		if reference, ok := resource.GetDependencyRun(rcName); runConfiguration.Status.LatestRuns.Succeeded.ProviderId != "" && !ok || reference.ProviderId != runConfiguration.Status.LatestRuns.Succeeded.ProviderId {
			resource.SetDependencyRun(rcName, runConfiguration.Status.LatestRuns.Succeeded)
			if err := dr.EC.Client.Status().Update(ctx, resource); err != nil {
				logger.Error(err, "error updating resource with observed pipeline version")
				return false, err
			}

			return true, nil
		}
	}

	return false, nil
}

func (dr DependingOnRunConfigurationReconciler[R]) getIgnoreNotFound(ctx context.Context, key client.ObjectKey) (*pipelinesv1.RunConfiguration, error) {
	logger := log.FromContext(ctx)
	pipeline := &pipelinesv1.RunConfiguration{}

	if err := dr.EC.Client.NonCached.Get(ctx, key, pipeline); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("object not found")
			return nil, nil
		}

		logger.Error(err, "unable to fetch object")
		return nil, err
	}

	return pipeline, nil
}

func (dr DependingOnRunConfigurationReconciler[R]) setupWithManager(mgr ctrl.Manager, controllerBuilder *builder.Builder, object client.Object, reconciliationRequestsForPipeline func(client.Object) []reconcile.Request) (*builder.Builder, error) {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), object, rcRefField, func(rawObj client.Object) []string {
		return rawObj.(R).GetReferencedDependencies()
	}); err != nil {
		return nil, err
	}

	return controllerBuilder.Watches(
		&source.Kind{Type: &pipelinesv1.RunConfiguration{}},
		handler.EnqueueRequestsFromMapFunc(reconciliationRequestsForPipeline),
		builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
	), nil
}
