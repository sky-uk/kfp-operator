package pipelines

import (
	"context"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/controllerconfigutil"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"time"

	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/logkeys"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowfactory"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

// ExperimentReconciler reconciles a Experiment object
type ExperimentReconciler struct {
	StateHandler[*pipelineshub.Experiment]
	ResourceReconciler[*pipelineshub.Experiment]
	ServiceManager ServiceResourceManager
}

func NewExperimentReconciler(
	ec K8sExecutionContext,
	workflowRepository WorkflowRepository,
	config config.KfpControllerConfigSpec,
) *ExperimentReconciler {
	return &ExperimentReconciler{
		StateHandler: StateHandler[*pipelineshub.Experiment]{
			WorkflowRepository: workflowRepository,
			WorkflowFactory:    workflowfactory.ExperimentWorkflowFactory(config),
		},
		ResourceReconciler: ResourceReconciler[*pipelineshub.Experiment]{
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

//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=experiments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=experiments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=experiments/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *ExperimentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()
	logger.V(2).Info("reconciliation started")

	var experiment = &pipelineshub.Experiment{}
	if err := r.EC.Client.NonCached.Get(ctx, req.NamespacedName, experiment); err != nil {
		logger.Error(err, "unable to fetch experiment")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(3).Info("found experiment", "resource", experiment)

	provider, err := r.LoadProvider(ctx, experiment.Spec.Provider)
	if err != nil {
		return ctrl.Result{}, err
	}

	providerSvc, err := r.ServiceManager.Get(ctx, &provider)
	if err != nil {
		return ctrl.Result{}, err
	}

	commands := r.StateHandler.StateTransition(ctx, provider, *providerSvc, experiment)

	for i := range commands {
		if err := commands[i].execute(ctx, r.EC, experiment); err != nil {
			logger.Error(err, "error executing command", logkeys.Command, commands[i])
			return ctrl.Result{}, err
		}
	}

	duration := time.Since(startTime)
	logger.V(2).Info("reconciliation ended", logkeys.Duration, duration)

	return ctrl.Result{}, nil
}

func (r *ExperimentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	experiment := &pipelineshub.Experiment{}
	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		For(experiment).
		WithOptions(controller.Options{
			RateLimiter: controllerconfigutil.RateLimiter,
		})

	controllerBuilder = r.ResourceReconciler.setupWithManager(controllerBuilder, experiment)

	return controllerBuilder.Complete(r)
}
