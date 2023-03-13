package pipelines

import (
	"context"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	pipelineRefField = ".spec.pipeline"
)

type DependingOnPipelineReconciler[R pipelinesv1.DependingOnPipelineResource] struct {
	BaseReconciler[R]
}

func (dr DependingOnPipelineReconciler[R]) handleObservedPipelineVersion(ctx context.Context, dependencyIdentifier pipelinesv1.PipelineIdentifier, resource R) error {
	logger := log.FromContext(ctx)

	setVersion := true
	desiredVersion := dependencyIdentifier.Version

	if dependencyIdentifier.Version == "" {
		pipeline := &pipelinesv1.Pipeline{}
		err := dr.getIgnoreNotFound(ctx, types.NamespacedName{
			Namespace: resource.GetNamespace(),
			Name:      dependencyIdentifier.Name,
		}, pipeline)
		if err != nil {
			return err
		}

		setVersion, desiredVersion = dependentPipelineVersionIfSucceeded(pipeline)
	}

	if setVersion && resource.GetObservedPipelineVersion() != desiredVersion {
		resource.SetObservedPipelineVersion(desiredVersion)

		if err := dr.EC.Client.Status().Update(ctx, resource); err != nil {
			logger.Error(err, "error updating resource with observed pipeline version")
			return err
		}
	}

	return nil
}

func (dr DependingOnPipelineReconciler[R]) getIgnoreNotFound(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	logger := log.FromContext(ctx)

	if err := dr.EC.Client.NonCached.Get(ctx, key, obj); !errors.IsNotFound(err) {
		logger.Error(err, "unable to fetch object")
		return err

	} else if err != nil {
		logger.Info("object not found")
	}

	return nil
}

func dependentPipelineVersionIfSucceeded(pipeline *pipelinesv1.Pipeline) (bool, string) {
	if pipeline == nil {
		return true, ""
	} else {
		switch pipeline.Status.SynchronizationState {
		case apis.Succeeded:
			return true, pipeline.Status.Version
		case apis.Deleted:
			return true, ""
		default:
			return false, ""
		}
	}
}
