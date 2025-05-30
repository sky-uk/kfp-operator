package pipelines

import (
	"context"

	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	pipelineRefField = ".spec.pipeline"
)

type DependingOnPipelineResource interface {
	client.Object
	GetPipeline() pipelineshub.PipelineIdentifier
	GetObservedPipelineVersion() string
	SetObservedPipelineVersion(string)
}

type DependingOnPipelineReconciler[R DependingOnPipelineResource] struct {
	EC K8sExecutionContext
}

func (dr DependingOnPipelineReconciler[R]) handleObservedPipelineVersion(ctx context.Context, pipelineIdentifier pipelineshub.PipelineIdentifier, resource R) (bool, error) {
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

func (dr DependingOnPipelineReconciler[R]) getIgnoreNotFound(ctx context.Context, key client.ObjectKey) (*pipelineshub.Pipeline, error) {
	logger := log.FromContext(ctx)
	pipeline := &pipelineshub.Pipeline{}

	if err := dr.EC.Client.NonCached.Get(ctx, key, pipeline); err != nil {
		if errors.IsNotFound(err) {
			logger.V(2).Info("object not found")
			return nil, nil
		}

		logger.Error(err, "unable to fetch object")
		return nil, err
	}

	return pipeline, nil
}

func (dr DependingOnPipelineReconciler[R]) setupWithManager(mgr ctrl.Manager, controllerBuilder *builder.Builder, object client.Object, reconciliationRequestsForPipeline handler.MapFunc) (*builder.Builder, error) {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), object, pipelineRefField, func(rawObj client.Object) []string {
		referencingResource := rawObj.(R)
		return []string{referencingResource.GetPipeline().Name}
	}); err != nil {
		return nil, err
	}

	return controllerBuilder.Watches(
		&pipelineshub.Pipeline{},
		handler.EnqueueRequestsFromMapFunc(reconciliationRequestsForPipeline),
		builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
	), nil
}

func dependentPipelineVersionIfSucceeded(pipeline *pipelineshub.Pipeline) (string, bool) {
	if pipeline == nil {
		return "", true
	}

	switch pipeline.Status.Conditions.GetSyncStateFromReason() {
	case apis.Succeeded:
		return pipeline.Status.Version, true
	case apis.Deleted:
		return "", true
	default:
		return "", false
	}
}
