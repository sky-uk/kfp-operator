package pipelines

import (
	"context"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	"github.com/google/go-cmp/cmp"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/logkeys"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowfactory"
	ctrl "sigs.k8s.io/controller-runtime"
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
	logger.Info("pipeline reconciliation started", "p", req.String())

	var pipeline = &pipelineshub.Pipeline{}
	if err := r.EC.Client.NonCached.Get(ctx, req.NamespacedName, pipeline); err != nil {
		logger.Error(err, "error loading pipeline", "pipeline", req.NamespacedName)
		return ctrl.Result{RequeueAfter: 10 * time.Minute}, nil
	}

	provider, err := r.LoadProvider(ctx, pipeline.Spec.Provider)
	if err != nil {
		logger.Error(err, "error loading provider", "prov", pipeline.Spec.Provider)
		return ctrl.Result{RequeueAfter: 10 * time.Minute}, nil
	}

	providerSvc, err := r.ServiceManager.Get(ctx, &provider)
	if err != nil {
		logger.Error(err, "error fetching provider service", "svc", provider.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Minute}, nil
	}

	commands := r.StateHandler.StateTransition(ctx, provider, *providerSvc, pipeline)

	for i := range commands {
		if err := commands[i].execute(ctx, r.EC, pipeline); err != nil {
			logger.Error(err, "error executing command", logkeys.Command, commands[i])
			return ctrl.Result{RequeueAfter: 10 * time.Minute}, nil
		}
	}

	duration := time.Since(startTime)

	logger.Info("reconciliation ended", logkeys.Duration, duration)
	return ctrl.Result{}, nil
}

func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pipeline := &pipelineshub.Pipeline{}
	logger := log.FromContext(ctx)

	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		For(pipeline, builder.WithPredicates(
			predicate.GenerationChangedPredicate{},
		)).WithEventFilter(predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			logger.Info("Create event", "name", e.Object.GetName())
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {

			oldObj := e.ObjectOld.DeepCopyObject().(client.Object)
			newObj := e.ObjectNew.DeepCopyObject().(client.Object)
			logger.Info("Update event", "name", e.ObjectNew.GetName(), diff := cmp.Diff(oldObj, newObj))
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			logger.Info("Delete event", "name", e.Object.GetName())
			return true
		},
	})

	controllerBuilder = r.ResourceReconciler.setupWithManager(controllerBuilder, pipeline)

	return controllerBuilder.Complete(r)
}
