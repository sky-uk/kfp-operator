package pipelines

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

var (
	workflowOwnerKey = ".metadata.controller"
	apiGVStr         = pipelinesv1.GroupVersion.String()
	finalizerName    = "finalizer.pipelines.kubeflow.org"
)

type PipelineReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	StateHandler PipelineStateHandler
}

//+kubebuilder:rbac:groups=argoproj.io,resources=workflows,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=pipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=pipelines/finalizers,verbs=update

func (r *PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var pipeline = &pipelinesv1.Pipeline{}
	if err := r.Get(ctx, req.NamespacedName, pipeline); err != nil {
		logger.Error(err, "unable to fetch pipeline")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if pipeline.ObjectMeta.DeletionTimestamp.IsZero() {
		r.AddFinalizer(ctx, pipeline)
	}

	commands := r.StateHandler.StateTransition(ctx, pipeline)

	for i := range commands {
		if err := commands[i].execute(r, ctx, pipeline); err != nil {
			logger.Error(err, fmt.Sprintf("Error executing command: %+v", commands[i]))
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *PipelineReconciler) AddFinalizer(ctx context.Context, pipeline *pipelinesv1.Pipeline) error {
	if !containsString(pipeline.ObjectMeta.Finalizers, finalizerName) {
		pipeline.ObjectMeta.Finalizers = append(pipeline.ObjectMeta.Finalizers, finalizerName)
		return r.Update(ctx, pipeline)
	}

	return nil
}

func (r *PipelineReconciler) RemoveFinalizer(ctx context.Context, pipeline *pipelinesv1.Pipeline) error {
	if containsString(pipeline.ObjectMeta.Finalizers, finalizerName) {
		pipeline.ObjectMeta.Finalizers = removeString(pipeline.ObjectMeta.Finalizers, finalizerName)
		return r.Update(ctx, pipeline)
	}

	return nil
}

func (r *PipelineReconciler) CreateChildWorkflow(ctx context.Context, pipeline *pipelinesv1.Pipeline, workflow argo.Workflow) error {
	if err := ctrl.SetControllerReference(pipeline, &workflow, r.Scheme); err != nil {
		return err
	}

	if err := r.Create(ctx, &workflow); err != nil {
		return err
	}

	return nil
}

func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pipelinesv1.Pipeline{}).
		Owns(&argo.Workflow{}).
		Complete(r)
}
