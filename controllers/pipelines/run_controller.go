package pipelines

import (
	"context"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RunReconciler reconciles a Run object
type RunReconciler struct {
	EC K8sExecutionContext
	StateHandler[*pipelinesv1.Run]
	DependingOnPipelineReconciler[*pipelinesv1.Run]
	DependingOnRunConfigurationReconciler[*pipelinesv1.Run]
	ResourceReconciler[*pipelinesv1.Run]
}

func NewRunReconciler(ec K8sExecutionContext, workflowRepository WorkflowRepository, config config.Configuration) *RunReconciler {
	return &RunReconciler{
		StateHandler: StateHandler[*pipelinesv1.Run]{
			WorkflowRepository: workflowRepository,
			WorkflowFactory:    RunWorkflowFactory(config),
		},
		EC: ec,
		DependingOnPipelineReconciler: DependingOnPipelineReconciler[*pipelinesv1.Run]{
			EC: ec,
		},
		DependingOnRunConfigurationReconciler: DependingOnRunConfigurationReconciler[*pipelinesv1.Run]{
			EC: ec,
		},
		ResourceReconciler: ResourceReconciler[*pipelinesv1.Run]{
			EC:     ec,
			Config: config,
		},
	}
}

//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=pipelines,verbs=get;list;watch

func (r *RunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()
	logger.V(2).Info("reconciliation started")

	var run = &pipelinesv1.Run{}
	if err := r.EC.Client.NonCached.Get(ctx, req.NamespacedName, run); err != nil {
		logger.Error(err, "unable to fetch run")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(3).Info("found run", "resource", run)

	result, err := r.handleCompletion(ctx, run)
	if err != nil {
		return result, err
	}

	desiredProvider := desiredProvider(run, r.Config)

	// Never change after being set
	if run.Status.ObservedPipelineVersion == "" || run.Spec.HasUnmetDependencies(run.Status.Dependencies) {
		if hasChanged, err := r.handleDependentRuns(ctx, run); hasChanged || err != nil {
			return ctrl.Result{}, err
		}

		if hasChanged, err := r.handleObservedPipelineVersion(ctx, run.Spec.Pipeline, run); hasChanged || err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	commands := r.StateHandler.StateTransition(ctx, desiredProvider, run)

	for i := range commands {
		if err := commands[i].execute(ctx, r.EC, run); err != nil {
			logger.Error(err, "error executing command", LogKeys.Command, commands[i])
			return result, err
		}
	}

	duration := time.Now().Sub(startTime)
	logger.V(2).Info("reconciliation ended", LogKeys.Duration, duration)

	return result, nil
}

func (r *RunReconciler) handleCompletion(ctx context.Context, run *pipelinesv1.Run) (ctrl.Result, error) {
	if err := r.markCompletedIfCompleted(ctx, run); err != nil {
		return ctrl.Result{}, err
	}

	if r.Config.RunCompletionTTL == nil || run.Status.MarkedCompletedAt == nil || run.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	ttlExpiry := run.Status.MarkedCompletedAt.Time.Add(r.Config.RunCompletionTTL.Duration)

	if time.Now().After(ttlExpiry) {
		err := r.EC.Client.Delete(ctx, run)
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Until(ttlExpiry)}, nil
}

func (r *RunReconciler) markCompletedIfCompleted(ctx context.Context, run *pipelinesv1.Run) error {
	if run.Status.CompletionState != "" && run.Status.MarkedCompletedAt == nil {
		now := metav1.Now()
		run.Status.MarkedCompletedAt = &now
		return r.EC.Client.Status().Update(ctx, run)
	}

	return nil
}

func (r *RunReconciler) reconciliationRequestsForPipeline(pipeline client.Object) []reconcile.Request {
	referencingRuns := &pipelinesv1.RunList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(pipelineRefField, pipeline.GetName()),
		Namespace:     pipeline.GetNamespace(),
	}

	err := r.EC.Client.Cached.List(context.TODO(), referencingRuns, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(referencingRuns.Items))
	for i, item := range referencingRuns.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

func (r *RunReconciler) reconciliationRequestsForRunconfigurations(runConfiguration client.Object) []reconcile.Request {
	referencingRuns := &pipelinesv1.RunList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(rcRefField, runConfiguration.GetName()),
		Namespace:     runConfiguration.GetNamespace(),
	}

	err := r.EC.Client.Cached.List(context.TODO(), referencingRuns, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(referencingRuns.Items))
	for i, item := range referencingRuns.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

func (r *RunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	run := &pipelinesv1.Run{}
	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		For(run)

	controllerBuilder = r.ResourceReconciler.setupWithManager(controllerBuilder, run)
	controllerBuilder, err := r.DependingOnPipelineReconciler.setupWithManager(mgr, controllerBuilder, run, r.reconciliationRequestsForPipeline)
	if err != nil {
		return err
	}
	controllerBuilder, err = r.DependingOnRunConfigurationReconciler.setupWithManager(mgr, controllerBuilder, run, r.reconciliationRequestsForRunconfigurations)
	if err != nil {
		return err
	}

	return controllerBuilder.Complete(r)
}
