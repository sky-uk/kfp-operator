package pipelines

import (
	"context"
	"time"

	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RunScheduleReconciler reconciles a RunSchedule object
type RunScheduleReconciler struct {
	StateHandler[*pipelinesv1.RunSchedule]
	ResourceReconciler[*pipelinesv1.RunSchedule]
}

func NewRunScheduleReconciler(ec K8sExecutionContext, workflowRepository WorkflowRepository, config config.Configuration) *RunScheduleReconciler {
	return &RunScheduleReconciler{
		StateHandler: StateHandler[*pipelinesv1.RunSchedule]{
			WorkflowRepository: workflowRepository,
			WorkflowFactory:    RunScheduleWorkflowFactory(config),
		},
		ResourceReconciler: ResourceReconciler[*pipelinesv1.RunSchedule]{
			EC:     ec,
			Config: config,
		},
	}
}

//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runschedules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runschedules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runschedules/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *RunScheduleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()
	logger.V(2).Info("reconciliation started")

	var runSchedule = &pipelinesv1.RunSchedule{}
	if err := r.EC.Client.NonCached.Get(ctx, req.NamespacedName, runSchedule); err != nil {
		logger.Error(err, "unable to fetch run schedule")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(3).Info("found run schedule", "resource", runSchedule)

	desiredProvider := desiredProvider(runSchedule, r.Config)

	provider := pipelinesv1.Provider{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: desiredProvider,
		},
		Spec:   pipelinesv1.ProviderSpec{},
		Status: pipelinesv1.ProviderStatus{},
	}

	commands := r.StateHandler.StateTransition(ctx, provider, runSchedule)

	for i := range commands {
		if err := commands[i].execute(ctx, r.EC, runSchedule); err != nil {
			logger.Error(err, "error executing command", LogKeys.Command, commands[i])
			return ctrl.Result{}, err
		}
	}

	duration := time.Now().Sub(startTime)
	logger.V(2).Info("reconciliation ended", LogKeys.Duration, duration)

	return ctrl.Result{}, nil
}

func (r *RunScheduleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	runSchedule := &pipelinesv1.RunSchedule{}
	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		For(runSchedule)

	controllerBuilder = r.ResourceReconciler.setupWithManager(controllerBuilder, runSchedule)

	return controllerBuilder.Complete(r)
}
