package pipelines

import (
	"context"
	"time"

	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/logkeys"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowfactory"
	"github.com/sky-uk/kfp-operator/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RunReconciler reconciles a Run object
type RunReconciler struct {
	EC K8sExecutionContext
	StateHandler[*pipelineshub.Run]
	DependingOnPipelineReconciler[*pipelineshub.Run]
	DependingOnRunConfigurationReconciler[*pipelineshub.Run]
	ResourceReconciler[*pipelineshub.Run]
	ServiceManager ServiceResourceManager
}

func NewRunReconciler(
	ec K8sExecutionContext,
	workflowRepository WorkflowRepository,
	config config.KfpControllerConfigSpec,
) *RunReconciler {
	return &RunReconciler{
		StateHandler: StateHandler[*pipelineshub.Run]{
			WorkflowRepository: workflowRepository,
			WorkflowFactory:    workflowfactory.RunWorkflowFactory(config),
		},
		EC: ec,
		DependingOnPipelineReconciler: DependingOnPipelineReconciler[*pipelineshub.Run]{
			EC: ec,
		},
		DependingOnRunConfigurationReconciler: DependingOnRunConfigurationReconciler[*pipelineshub.Run]{
			EC: ec,
		},
		ResourceReconciler: ResourceReconciler[*pipelineshub.Run]{
			EC:     ec,
			Config: config,
		},
		ServiceManager: ServiceManager{
			client: &ec.Client,
			scheme: ec.Scheme,
			config: &config,
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

	var run = &pipelineshub.Run{}
	if err := r.EC.Client.NonCached.Get(ctx, req.NamespacedName, run); err != nil {
		logger.Error(err, "unable to fetch run")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(3).Info("found run", "resource", run)

	result, err := r.handleCompletion(ctx, run)
	if err != nil {
		return result, err
	}

	provider, err := r.LoadProvider(ctx, run.Spec.Provider)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Never change after being set
	if run.Status.Dependencies.Pipeline.Version == "" || run.Spec.HasUnmetDependencies(run.Status.Dependencies) {
		if hasChanged, err := r.handleDependentRuns(ctx, run); hasChanged || err != nil {
			return ctrl.Result{}, err
		}

		if hasChanged, err := r.handleObservedPipelineVersion(ctx, run.Spec.Pipeline, run); hasChanged || err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	providerSvc, err := r.ServiceManager.Get(ctx, &provider)
	if err != nil {
		return ctrl.Result{}, err
	}

	commands := r.StateHandler.StateTransition(ctx, provider, *providerSvc, run)

	for i := range commands {
		if err := commands[i].execute(ctx, r.EC, run); err != nil {
			logger.Error(err, "error executing command", logkeys.Command, commands[i])
			return result, err
		}
	}

	duration := time.Now().Sub(startTime)
	logger.V(2).Info("reconciliation ended", logkeys.Duration, duration)

	return result, nil
}

func (r *RunReconciler) handleCompletion(ctx context.Context, run *pipelineshub.Run) (ctrl.Result, error) {
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

func (r *RunReconciler) markCompletedIfCompleted(ctx context.Context, run *pipelineshub.Run) error {
	if run.Status.CompletionState != "" && run.Status.MarkedCompletedAt == nil {
		now := metav1.Now()
		run.Status.MarkedCompletedAt = &now
		return r.EC.Client.Status().Update(ctx, run)
	}

	return nil
}

func (r *RunReconciler) reconciliationRequestsForPipeline(
	ctx context.Context,
	pipeline client.Object,
) []reconcile.Request {
	referencingRuns := &pipelineshub.RunList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(pipelineRefField, pipeline.GetName()),
		Namespace:     pipeline.GetNamespace(),
	}

	err := r.EC.Client.Cached.List(ctx, referencingRuns, listOps)
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

func (r *RunReconciler) reconciliationRequestsForRunconfigurations(
	ctx context.Context,
	runConfiguration client.Object,
) []reconcile.Request {
	referencingRuns := &pipelineshub.RunList{}
	rcNamespacedName, err := common.NamespacedName{
		Name:      runConfiguration.GetName(),
		Namespace: runConfiguration.GetNamespace(),
	}.String()
	if err != nil {
		return []reconcile.Request{}
	}

	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(rcRefField, rcNamespacedName),
	}

	err = r.EC.Client.Cached.List(ctx, referencingRuns, listOps)
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
	run := &pipelineshub.Run{}
	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		For(run)

	controllerBuilder = r.ResourceReconciler.setupWithManager(controllerBuilder, run)
	controllerBuilder, err := r.DependingOnPipelineReconciler.setupWithManager(
		mgr,
		controllerBuilder,
		run,
		r.reconciliationRequestsForPipeline,
	)
	if err != nil {
		return err
	}
	controllerBuilder, err = r.DependingOnRunConfigurationReconciler.setupWithManager(
		mgr,
		controllerBuilder,
		run,
		r.reconciliationRequestsForRunconfigurations,
	)
	if err != nil {
		return err
	}

	return controllerBuilder.Complete(r)
}
