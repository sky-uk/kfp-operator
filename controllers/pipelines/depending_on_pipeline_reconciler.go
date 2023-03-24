package pipelines

import (
	"context"
	"github.com/sky-uk/kfp-operator/apis"
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
	pipelineRefField = ".spec.pipeline"
)

type DependingOnPipelineResource interface {
	client.Object
	//pipelinesv1.Resource
	GetPipeline() pipelinesv1.PipelineIdentifier
	GetObservedPipelineVersion() string
	SetObservedPipelineVersion(string)
}

type DependingOnPipelineReconciler[R DependingOnPipelineResource] struct {
	EC K8sExecutionContext
}

func (dr DependingOnPipelineReconciler[R]) handleObservedPipelineVersion(ctx context.Context, pipelineIdentifier pipelinesv1.PipelineIdentifier, resource R) (bool, error) {
	logger := log.FromContext(ctx)

	setVersion := true
	desiredVersion := pipelineIdentifier.Version

	if pipelineIdentifier.Version == "" {
		pipeline, err := dr.getIgnoreNotFound(ctx, types.NamespacedName{
			Namespace: resource.GetNamespace(),
			Name:      pipelineIdentifier.Name,
		})
		if err != nil {
			return false, err
		}

		desiredVersion, setVersion = dependentPipelineVersionIfSucceeded(pipeline)
	}

	if setVersion && resource.GetObservedPipelineVersion() != desiredVersion {
		resource.SetObservedPipelineVersion(desiredVersion)

		if err := dr.EC.Client.Status().Update(ctx, resource); err != nil {
			logger.Error(err, "error updating resource with observed pipeline version")
			return false, err
		}

		return true, nil
	}

	return false, nil
}

func (dr DependingOnPipelineReconciler[R]) getIgnoreNotFound(ctx context.Context, key client.ObjectKey) (*pipelinesv1.Pipeline, error) {
	logger := log.FromContext(ctx)
	pipeline := &pipelinesv1.Pipeline{}

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

func (dr DependingOnPipelineReconciler[R]) setupWithManager(mgr ctrl.Manager, controllerBuilder *builder.Builder, object client.Object, reconciliationRequestsForPipeline func(client.Object) []reconcile.Request) (*builder.Builder, error) {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), object, pipelineRefField, func(rawObj client.Object) []string {
		referencingResource := rawObj.(R)
		return []string{referencingResource.GetPipeline().Name}
	}); err != nil {
		return nil, err
	}

	return controllerBuilder.Watches(
		&source.Kind{Type: &pipelinesv1.Pipeline{}},
		handler.EnqueueRequestsFromMapFunc(reconciliationRequestsForPipeline),
		builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
	), nil
}

func dependentPipelineVersionIfSucceeded(pipeline *pipelinesv1.Pipeline) (string, bool) {
	if pipeline == nil {
		return "", true
	}

	switch pipeline.Status.SynchronizationState {
	case apis.Succeeded:
		return pipeline.Status.Version, true
	case apis.Deleted:
		return "", true
	default:
		return "", false
	}
}
