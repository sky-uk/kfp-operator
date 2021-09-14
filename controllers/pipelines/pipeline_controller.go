package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	pipelineWorkflows "github.com/sky-uk/kfp-operator/controllers/pipelines/workflows"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	workflowOwnerKey = ".metadata.controller"
	apiGVStr         = pipelinesv1.GroupVersion.String()
	finalizerName    = "finalizer.pipelines.kubeflow.org"
)

type PipelineReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Workflows pipelineWorkflows.Workflows
}

type Command interface {
	execute(*PipelineReconciler, context.Context, *pipelinesv1.Pipeline) error
}

type SetPipelineStatus struct {
	Status pipelinesv1.PipelineStatus
}

func (sps SetPipelineStatus) execute(reconciler *PipelineReconciler, ctx context.Context, pipeline *pipelinesv1.Pipeline) error {
	pipeline.Status = sps.Status

	return reconciler.Status().Update(ctx, pipeline)
}

type CreateWorkflow struct {
	Workflow argo.Workflow
}

func (cw CreateWorkflow) execute(reconciler *PipelineReconciler, ctx context.Context, pipeline *pipelinesv1.Pipeline) error {
	return reconciler.CreateChildWorkflow(ctx, pipeline, cw.Workflow)
}

type DeleteWorkflows struct {
	Workflows []argo.Workflow
}

func (dw DeleteWorkflows) execute(reconciler *PipelineReconciler, ctx context.Context, pipeline *pipelinesv1.Pipeline) error {
	for i := range dw.Workflows {
		if err := reconciler.Delete(ctx, &dw.Workflows[i]); err != nil {
			return err
		}
	}

	return nil
}

type DeletePipeline struct {
}

func (dp DeletePipeline) execute(reconciler *PipelineReconciler, ctx context.Context, pipeline *pipelinesv1.Pipeline) error {
	return reconciler.RemoveFinalizer(ctx, *pipeline)
}

type WorkflowsProvider interface {
	GetByOperation(operation string) []argo.Workflow
}

type WorkflowsImpl struct {
	*PipelineReconciler
	ctx      context.Context
	pipeline *pipelinesv1.Pipeline
}

func (w WorkflowsImpl) GetByOperation(operation string) []argo.Workflow {
	var workflows argo.WorkflowList

	w.List(w.ctx, &workflows, client.InNamespace(w.pipeline.ObjectMeta.Namespace), client.MatchingLabels{pipelineWorkflows.OperationLabelKey: operation, pipelineWorkflows.PipelineLabelKey: w.pipeline.ObjectMeta.Name})

	return workflows.Items
}

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

	workflows := WorkflowsImpl{r, ctx, pipeline}
	stateHandler := StateHandler{
		Workflows: r.Workflows,
	}
	commands := stateHandler.StateTransition(pipeline, workflows)

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

func (r *PipelineReconciler) RemoveFinalizer(ctx context.Context, pipeline pipelinesv1.Pipeline) error {
	if containsString(pipeline.ObjectMeta.Finalizers, finalizerName) {
		pipeline.ObjectMeta.Finalizers = removeString(pipeline.ObjectMeta.Finalizers, finalizerName)
		return r.Update(ctx, &pipeline)
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

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &argo.Workflow{}, workflowOwnerKey, func(rawObj client.Object) []string {
		workflow := rawObj.(*argo.Workflow)

		owner := metav1.GetControllerOf(workflow)

		if owner == nil {
			return nil
		}

		if owner.APIVersion != apiGVStr || owner.Kind != "Pipeline" {
			return nil
		}

		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&pipelinesv1.Pipeline{}).
		Owns(&argo.Workflow{}).
		Complete(r)
}
