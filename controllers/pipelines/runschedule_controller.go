package pipelines

import (
	"context"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/controllerconfigutil"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"time"

	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/logkeys"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowfactory"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RunScheduleReconciler reconciles a RunSchedule object
type RunScheduleReconciler struct {
	StateHandler[*pipelineshub.RunSchedule]
	ResourceReconciler[*pipelineshub.RunSchedule]
	ServiceManager ServiceResourceManager
}

func NewRunScheduleReconciler(
	ec K8sExecutionContext,
	workflowRepository WorkflowRepository,
	config config.KfpControllerConfigSpec,
) *RunScheduleReconciler {
	return &RunScheduleReconciler{
		StateHandler: StateHandler[*pipelineshub.RunSchedule]{
			WorkflowRepository: workflowRepository,
			WorkflowFactory:    workflowfactory.RunScheduleWorkflowFactory(config),
		},
		ResourceReconciler: ResourceReconciler[*pipelineshub.RunSchedule]{
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

//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runschedules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runschedules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runschedules/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *RunScheduleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()
	logger.V(2).Info("reconciliation started")

	var runSchedule = &pipelineshub.RunSchedule{}
	if err := r.EC.Client.NonCached.Get(ctx, req.NamespacedName, runSchedule); err != nil {
		logger.Error(err, "unable to fetch run schedule")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(3).Info("found run schedule", "resource", runSchedule)

	provider, err := r.LoadProvider(ctx, runSchedule.Spec.Provider)
	if err != nil {
		logger.Error(err, "Failed to load provider", "provider", provider)
		return ctrl.Result{}, err
	}

	providerSvc, err := r.ServiceManager.Get(ctx, &provider)
	if err != nil {
		return ctrl.Result{}, err
	}

	commands := r.StateHandler.StateTransition(ctx, provider, *providerSvc, runSchedule)

	for i := range commands {
		if err := commands[i].execute(ctx, r.EC, runSchedule); err != nil {
			logger.Error(err, "error executing command", logkeys.Command, commands[i])
			return ctrl.Result{}, err
		}
	}

	duration := time.Since(startTime)
	logger.V(2).Info("reconciliation ended", logkeys.Duration, duration)

	return ctrl.Result{}, nil
}

func (r *RunScheduleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	runSchedule := &pipelineshub.RunSchedule{}
	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		For(runSchedule).
		WithOptions(controller.Options{
			RateLimiter: controllerconfigutil.RateLimiter,
		})

	controllerBuilder = r.ResourceReconciler.setupWithManager(controllerBuilder, runSchedule)

	return controllerBuilder.Complete(r)
}
