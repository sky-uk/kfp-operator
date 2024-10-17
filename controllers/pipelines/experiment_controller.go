package pipelines

import (
	"context"
	"time"

	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
)

// ExperimentReconciler reconciles a Experiment object
type ExperimentReconciler struct {
	StateHandler[*pipelinesv1.Experiment]
	ResourceReconciler[*pipelinesv1.Experiment]
}

func NewExperimentReconciler(ec K8sExecutionContext, workflowRepository WorkflowRepository, config config.KfpControllerConfigSpec) *ExperimentReconciler {
	return &ExperimentReconciler{
		StateHandler: StateHandler[*pipelinesv1.Experiment]{
			WorkflowRepository: workflowRepository,
			WorkflowFactory:    ExperimentWorkflowFactory(config),
		},
		ResourceReconciler: ResourceReconciler[*pipelinesv1.Experiment]{
			EC:     ec,
			Config: config,
		},
	}
}

//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=experiments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=experiments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=experiments/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *ExperimentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()
	logger.V(2).Info("reconciliation started")

	var experiment = &pipelinesv1.Experiment{}
	if err := r.EC.Client.NonCached.Get(ctx, req.NamespacedName, experiment); err != nil {
		logger.Error(err, "unable to fetch experiment")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(3).Info("found experiment", "resource", experiment)

	desiredProvider := desiredProvider(experiment, r.Config)

	provider, err := r.loadProvider(ctx, r.Config.WorkflowNamespace, desiredProvider)
	if err != nil {
		return ctrl.Result{}, err
	}

	commands := r.StateHandler.StateTransition(ctx, provider, experiment)

	for i := range commands {
		if err := commands[i].execute(ctx, r.EC, experiment); err != nil {
			logger.Error(err, "error executing command", LogKeys.Command, commands[i])
			return ctrl.Result{}, err
		}
	}

	duration := time.Now().Sub(startTime)
	logger.V(2).Info("reconciliation ended", LogKeys.Duration, duration)

	return ctrl.Result{}, nil
}

func (r *ExperimentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	experiment := &pipelinesv1.Experiment{}
	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		For(experiment)

	controllerBuilder = r.ResourceReconciler.setupWithManager(controllerBuilder, experiment)

	return controllerBuilder.Complete(r)
}
