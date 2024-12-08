package pipelines

import (
	"context"
	"time"

	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/logging"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	workflowOwnerKey = ".metadata.controller"
	apiGVStr         = pipelinesv1.GroupVersion.String()
	finalizerName    = "finalizer.pipelines.kubeflow.org"
)

type PipelineReconciler struct {
	StateHandler[*pipelinesv1.Pipeline]
	ResourceReconciler[*pipelinesv1.Pipeline]
}

func NewPipelineReconciler(ec K8sExecutionContext, workflowRepository WorkflowRepository, config config.KfpControllerConfigSpec) *PipelineReconciler {
	return &PipelineReconciler{
		StateHandler: StateHandler[*pipelinesv1.Pipeline]{
			WorkflowRepository: workflowRepository,
			WorkflowFactory:    PipelineWorkflowFactory(config),
		},
		ResourceReconciler: ResourceReconciler[*pipelinesv1.Pipeline]{
			EC:     ec,
			Config: config,
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

	var pipeline = &pipelinesv1.Pipeline{}
	if err := r.EC.Client.NonCached.Get(ctx, req.NamespacedName, pipeline); err != nil {
		logger.Error(err, "unable to fetch pipeline")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(3).Info("found pipeline", "resource", pipeline)

	provider, err := r.loadProvider(ctx, r.Config.WorkflowNamespace, pipeline.Spec.Provider)
	if err != nil {
		return ctrl.Result{}, err
	}

	commands := r.StateHandler.StateTransition(ctx, provider, pipeline)

	for i := range commands {
		if err := commands[i].execute(ctx, r.EC, pipeline); err != nil {
			logger.Error(err, "error executing command", logging.LogKeys.Command, commands[i])
			return ctrl.Result{}, err
		}
	}

	duration := time.Now().Sub(startTime)
	logger.V(2).Info("reconciliation ended", logging.LogKeys.Duration, duration)

	return ctrl.Result{}, nil
}

func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pipeline := &pipelinesv1.Pipeline{}
	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		For(pipeline)

	controllerBuilder = r.ResourceReconciler.setupWithManager(controllerBuilder, pipeline)

	return controllerBuilder.Complete(r)
}
