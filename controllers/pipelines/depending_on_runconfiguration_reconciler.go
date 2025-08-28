package pipelines

import (
	"context"
	"github.com/samber/lo"
	"reflect"

	"github.com/sky-uk/kfp-operator/apis"

	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/pkg/common"
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
	rcRefField = ".referencedRunConfigurations"
)

type DependingOnRunConfigurationResource interface {
	client.Object
	GetReferencedRCs() []common.NamespacedName
	GetReferencedRCArtifacts() []pipelineshub.RunConfigurationRef
	GetDependencyRuns() map[string]pipelineshub.RunReference
	SetDependencyRuns(map[string]pipelineshub.RunReference)
}

type DependingOnRunConfigurationReconciler[R DependingOnRunConfigurationResource] struct {
	EC K8sExecutionContext
}

func (dr DependingOnRunConfigurationReconciler[R]) handleDependentRuns(ctx context.Context, resource R) (bool, error) {
	logger := log.FromContext(ctx)

	artifactReferencesByDependency := lo.GroupByMap(resource.GetReferencedRCArtifacts(), func(r pipelineshub.RunConfigurationRef) (common.NamespacedName, string) {
		return r.Name, r.OutputArtifact
	})
	for _, rc := range resource.GetReferencedRCs() {
		if _, ok := artifactReferencesByDependency[rc]; !ok {
			artifactReferencesByDependency[rc] = nil
		}
	}

	dependencies := make(map[string]pipelineshub.RunReference)

	for dependencyName, artifactReferences := range artifactReferencesByDependency {
		dependencyNamespacedName, err := dependencyName.String()
		if err != nil {
			return false, err
		}
		namespace := dependencyName.Namespace
		if dependencyName.Namespace == "" {
			namespace = resource.GetNamespace()
		}

		runConfiguration, err := dr.getIgnoreNotFound(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      dependencyName.Name,
		})
		if err != nil {
			return false, err
		}

		if runConfiguration == nil {
			dependencies[dependencyNamespacedName] = pipelineshub.RunReference{}
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

		dependencies[dependencyNamespacedName] = pipelineshub.RunReference{
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

func (dr DependingOnRunConfigurationReconciler[R]) getIgnoreNotFound(ctx context.Context, key client.ObjectKey) (*pipelineshub.RunConfiguration, error) {
	logger := log.FromContext(ctx)
	runConfiguration := &pipelineshub.RunConfiguration{}

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

func (dr DependingOnRunConfigurationReconciler[R]) setupWithManager(mgr ctrl.Manager, controllerBuilder *builder.Builder, resource client.Object, reconciliationRequestsForPipeline handler.MapFunc) (*builder.Builder, error) {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), resource, rcRefField, func(rawObj client.Object) []string {
		rcNamespacedNames, _ := apis.MapErr(rawObj.(R).GetReferencedRCs(), func(nsn common.NamespacedName) (string, error) {
			if nsn.Namespace == "" {
				nsn.Namespace = rawObj.GetNamespace()
			}
			return nsn.String()
		})
		return rcNamespacedNames
	}); err != nil {
		return nil, err
	}

	return controllerBuilder.Watches(
		&pipelineshub.RunConfiguration{},
		handler.EnqueueRequestsFromMapFunc(reconciliationRequestsForPipeline),
		builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
	), nil
}
