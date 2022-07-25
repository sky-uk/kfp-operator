package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha1"
)

// ExperimentReconciler reconciles a Experiment object
type ExperimentReconciler struct {
	EC           K8sExecutionContext
	StateHandler ExperimentStateHandler
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

	commands := r.StateHandler.StateTransition(ctx, experiment)

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
	return ctrl.NewControllerManagedBy(mgr).
		For(&pipelinesv1.Experiment{}).
		Owns(&argo.Workflow{}).
		Complete(r)
}
