package pipelines

import (
	"context"
	"net/http"
	"time"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha1"
)

const (
	pipelineRefField = ".spec.pipeline"
)

// RunConfigurationReconciler reconciles a RunConfiguration object
type RunConfigurationReconciler struct {
	EC           K8sExecutionContext
	StateHandler RunConfigurationStateHandler
}

//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runconfigurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runconfigurations/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=pipeline,verbs=get;list;watch

func (r *RunConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()
	logger.V(2).Info("reconciliation started")

	var runConfiguration = &pipelinesv1.RunConfiguration{}
	if err := r.EC.Client.NonCached.Get(ctx, req.NamespacedName, runConfiguration); err != nil {
		logger.Error(err, "unable to fetch run configuration")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(3).Info("found run configuration", "resource", runConfiguration)

	var desiredVersion *string
	_, version := runConfiguration.ExtractPipelineNameVersion()

	if version == "" {
		pipeline, err := r.fetchDependentPipeline(ctx, runConfiguration)
		if err != nil {
			return ctrl.Result{}, err
		}

		desiredVersion = dependentPipelineVersionIfStable(pipeline)
	} else {
		desiredVersion = &version
	}

	if desiredVersion != nil && *desiredVersion != runConfiguration.Status.ObservedPipelineVersion {
		runConfiguration.Status.ObservedPipelineVersion = *desiredVersion
		err := r.EC.Client.Status().Update(ctx, runConfiguration)

		if err != nil {
			logger.Error(err, "error updating run configuration with new desired pipeline version")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	commands := r.StateHandler.StateTransition(ctx, runConfiguration)

	for i := range commands {
		if err := commands[i].execute(ctx, r.EC, runConfiguration); err != nil {
			logger.Error(err, "error executing command", LogKeys.Command, commands[i])
			return ctrl.Result{}, err
		}
	}

	duration := time.Now().Sub(startTime)
	logger.V(2).Info("reconciliation ended", LogKeys.Duration, duration)

	return ctrl.Result{}, nil
}

func (r *RunConfigurationReconciler) fetchDependentPipeline(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) (*pipelinesv1.Pipeline, error) {
	logger := log.FromContext(ctx)

	if runConfiguration.Spec.Pipeline == "" {
		return nil, nil
	}

	pipeline := &pipelinesv1.Pipeline{}
	pipelineName, _ := runConfiguration.ExtractPipelineNameVersion()

	if err := r.EC.Client.NonCached.Get(ctx, types.NamespacedName{Namespace: runConfiguration.Namespace, Name: pipelineName}, pipeline); err != nil {
		if statusError, isStatusError := err.(*errors.StatusError); !isStatusError || statusError.ErrStatus.Code != http.StatusNotFound {
			logger.Error(err, "unable to fetch dependent pipeline")

			return nil, err
		}

		logger.Info("dependent pipeline not found")
		return nil, nil
	}

	return pipeline, nil
}

func dependentPipelineVersionIfStable(dependentPipeline *pipelinesv1.Pipeline) *string {
	empty := ""

	if dependentPipeline == nil {
		return &empty
	} else {
		switch dependentPipeline.Status.SynchronizationState {
		case pipelinesv1.Succeeded:
			return &dependentPipeline.Status.Version
		case pipelinesv1.Deleted:
			return &empty
		default:
			return nil
		}
	}
}

func (r *RunConfigurationReconciler) reconciliationRequestsForPipeline(pipeline client.Object) []reconcile.Request {
	referencingRunConfigurations := &pipelinesv1.RunConfigurationList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(pipelineRefField, pipeline.GetName()),
		Namespace:     pipeline.GetNamespace(),
	}
	err := r.EC.Client.Cached.List(context.TODO(), referencingRunConfigurations, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(referencingRunConfigurations.Items))
	for i, item := range referencingRunConfigurations.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

func (r *RunConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &pipelinesv1.RunConfiguration{}, pipelineRefField, func(rawObj client.Object) []string {
		runConfiguration := rawObj.(*pipelinesv1.RunConfiguration)
		if runConfiguration.Spec.Pipeline == "" {
			return nil
		}

		pipelineName, _ := runConfiguration.ExtractPipelineNameVersion()
		return []string{pipelineName}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&pipelinesv1.RunConfiguration{}).
		Owns(&argo.Workflow{}).
		Watches(
			&source.Kind{Type: &pipelinesv1.Pipeline{}},
			handler.EnqueueRequestsFromMapFunc(r.reconciliationRequestsForPipeline),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}
