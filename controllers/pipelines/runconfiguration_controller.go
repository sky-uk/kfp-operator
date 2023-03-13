package pipelines

import (
	"context"
	"time"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
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

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
)

// RunConfigurationReconciler reconciles a RunConfiguration object
type RunConfigurationReconciler struct {
	DependingOnPipelineReconciler[*pipelinesv1.RunConfiguration]
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

	desiredProvider := r.desiredProvider(runConfiguration)

	if err := r.handleObservedPipelineVersion(ctx, runConfiguration.Spec.Pipeline, runConfiguration); err != nil {
		return ctrl.Result{}, err
	}

	commands := r.StateHandler.StateTransition(ctx, desiredProvider, runConfiguration)

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

func (r *RunConfigurationReconciler) reconciliationRequestsWorkflow(workflow client.Object) []reconcile.Request {
	return r.BaseReconciler.reconciliationRequestsWorkflow(workflow, &pipelinesv1.RunConfiguration{})
}

func (r *RunConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &pipelinesv1.RunConfiguration{}, pipelineRefField, func(rawObj client.Object) []string {
		runConfiguration := rawObj.(*pipelinesv1.RunConfiguration)
		return []string{runConfiguration.Spec.Pipeline.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&pipelinesv1.RunConfiguration{}).
		Watches(
			&source.Kind{Type: &pipelinesv1.Pipeline{}},
			handler.EnqueueRequestsFromMapFunc(r.reconciliationRequestsForPipeline),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Watches(&source.Kind{Type: &argo.Workflow{}},
			handler.EnqueueRequestsFromMapFunc(r.reconciliationRequestsWorkflow),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}
