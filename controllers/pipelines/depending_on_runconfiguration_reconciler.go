package pipelines

import (
	"context"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
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
	GetReferencedDependencies() []pipelinesv1.RunConfigurationRef
	GetDependencyRuns() map[string]pipelinesv1.RunReference
	SetDependencyRuns(map[string]pipelinesv1.RunReference)
}

type DependingOnRunConfigurationReconciler[R DependingOnRunConfigurationResource] struct {
	EC K8sExecutionContext
}

func (dr DependingOnRunConfigurationReconciler[R]) handleDependentRuns(ctx context.Context, resource R) (bool, error) {
	logger := log.FromContext(ctx)

	artifactReferencesByDependency := pipelines.GroupMap(resource.GetReferencedDependencies(), func(r pipelinesv1.RunConfigurationRef) (string, string) {
		return r.Name, r.OutputArtifact
	})

	dependencies := make(map[string]pipelinesv1.RunReference)

	for dependencyName, artifactReferences := range artifactReferencesByDependency {
		runConfiguration, err := dr.getIgnoreNotFound(ctx, types.NamespacedName{
			Namespace: resource.GetNamespace(),
			Name:      dependencyName,
		})
		if err != nil {
			return false, err
		}

		if runConfiguration == nil {
			dependencies[dependencyName] = pipelinesv1.RunReference{}
			continue
		}

		var dependencyArtifacts []common.Artifact

		for _, artifactReference := range artifactReferences {
			for _, artifact := range runConfiguration.Status.LatestRuns.Succeeded.Artifacts {
				if artifactReference == artifact.Name {
					dependencyArtifacts = append(dependencyArtifacts, artifact)
					break
				}
			}
		}

		dependencies[dependencyName] = pipelinesv1.RunReference{
			ProviderId: runConfiguration.Status.LatestRuns.Succeeded.ProviderId,
			Artifacts:  dependencyArtifacts,
		}

	}

	if !reflect.DeepEqual(resource.GetDependencyRuns(), dependencies) {
		resource.SetDependencyRuns(dependencies)
		if err := dr.EC.Client.Status().Update(ctx, resource); err != nil {
			logger.Error(err, "error updating resource with dependencies")
			return false, err
		}

		return true, nil
	}

	return false, nil
}

func (dr DependingOnRunConfigurationReconciler[R]) getIgnoreNotFound(ctx context.Context, key client.ObjectKey) (*pipelinesv1.RunConfiguration, error) {
	logger := log.FromContext(ctx)
	runConfiguration := &pipelinesv1.RunConfiguration{}

	if err := dr.EC.Client.NonCached.Get(ctx, key, runConfiguration); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("object not found")
			return nil, nil
		}

		logger.Error(err, "unable to fetch object")
		return nil, err
	}

	return runConfiguration, nil
}

func (dr DependingOnRunConfigurationReconciler[R]) setupWithManager(mgr ctrl.Manager, controllerBuilder *builder.Builder, object client.Object, reconciliationRequestsForPipeline func(client.Object) []reconcile.Request) (*builder.Builder, error) {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), object, rcRefField, func(rawObj client.Object) []string {
		return pipelines.Map(rawObj.(R).GetReferencedDependencies(), func(r pipelinesv1.RunConfigurationRef) string {
			return r.Name
		})
	}); err != nil {
		return nil, err
	}

	return controllerBuilder.Watches(
		&source.Kind{Type: &pipelinesv1.RunConfiguration{}},
		handler.EnqueueRequestsFromMapFunc(reconciliationRequestsForPipeline),
		builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
	), nil
}
