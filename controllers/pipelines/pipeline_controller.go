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

var (
	workflowOwnerKey = ".metadata.controller"
	apiGVStr         = pipelineshub.GroupVersion.String()
	finalizerName    = "finalizer.pipelines.kubeflow.org"
)

type PipelineReconciler struct {
	StateHandler[*pipelineshub.Pipeline]
	ResourceReconciler[*pipelineshub.Pipeline]
	ServiceManager ServiceResourceManager
}

func NewPipelineReconciler(
	ec K8sExecutionContext,
	workflowRepository WorkflowRepository,
	config config.KfpControllerConfigSpec,
) *PipelineReconciler {
	return &PipelineReconciler{
		StateHandler: StateHandler[*pipelineshub.Pipeline]{
			WorkflowRepository: workflowRepository,
			WorkflowFactory:    workflowfactory.PipelineWorkflowFactory(config),
		},
		ResourceReconciler: ResourceReconciler[*pipelineshub.Pipeline]{
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

//+kubebuilder:rbac:groups=argoproj.io,resources=workflows,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=pipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=pipelines/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()
	logger.V(2).Info("reconciliation started")

	var pipeline = &pipelineshub.Pipeline{}
	if err := r.EC.Client.NonCached.Get(ctx, req.NamespacedName, pipeline); err != nil {
		logger.Error(err, "unable to fetch pipeline")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(3).Info("found pipeline", "resource", pipeline)

	provider, err := r.LoadProvider(ctx, pipeline.Spec.Provider)
	if err != nil {
		return ctrl.Result{}, err
	}

	providerSvc, err := r.ServiceManager.Get(ctx, &provider)
	if err != nil {
		return ctrl.Result{}, err
	}

	commands := r.StateHandler.StateTransition(ctx, provider, *providerSvc, pipeline)

	for i := range commands {
		if err := commands[i].execute(ctx, r.EC, pipeline); err != nil {
			logger.Error(err, "error executing command", logkeys.Command, commands[i])
			return ctrl.Result{}, err
		}
	}

	duration := time.Now().Sub(startTime)
	logger.V(2).Info("reconciliation ended", logkeys.Duration, duration)

	return ctrl.Result{}, nil
}

func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pipeline := &pipelineshub.Pipeline{}
	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		For(pipeline).
		WithOptions(controller.Options{
			RateLimiter: controllerconfigutil.RateLimiter,
		})

	controllerBuilder = r.ResourceReconciler.setupWithManager(controllerBuilder, pipeline)

	return controllerBuilder.Complete(r)
}
