package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	workflowOwnerKey = ".metadata.controller"
	apiGVStr         = pipelinesv1.GroupVersion.String()
)

type PipelineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=pipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=pipelines/finalizers,verbs=update

func (r *PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var pipeline pipelinesv1.Pipeline
	if err := r.Get(ctx, req.NamespacedName, &pipeline); err != nil {
		logger.Error(err, "unable to fetch pipeline")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	switch pipeline.Status.SynchronizationState {
	case pipelinesv1.Creating:
		if err := r.onCreationWorkflowSucceeded(ctx, pipeline); err != nil {
			logger.Error(err, "unable to create pipeline on Kubeflow")
			return ctrl.Result{}, err
		}
	case pipelinesv1.Unknown:
		if err := r.onCreation(ctx, pipeline); err != nil {
			logger.Error(err, "unable to create pipeline on Kubeflow")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *PipelineReconciler) onCreation(ctx context.Context, pipeline pipelinesv1.Pipeline) error {
	workflow := constructUploadWorkflow(&pipeline)

	if err := ctrl.SetControllerReference(&pipeline, workflow, r.Scheme); err != nil {
		return err
	}

	if err := r.Create(ctx, workflow); err != nil {
		return err
	}

	pipeline.Status.SynchronizationState = pipelinesv1.Creating

	if err := r.Status().Update(ctx, &pipeline); err != nil {
		return err
	}

	return nil
}

func (r *PipelineReconciler) onCreationWorkflowSucceeded(ctx context.Context, pipeline pipelinesv1.Pipeline) error {
	var childWorkflows argo.WorkflowList

	if err := r.List(ctx, &childWorkflows, client.InNamespace(pipeline.ObjectMeta.Namespace), client.MatchingFields{workflowOwnerKey: pipeline.ObjectMeta.Name}); err != nil {
		return err
	}

	if len(childWorkflows.Items) > 0 {
		workflow := childWorkflows.Items[0]

		switch workflow.Status.Phase {
		case argo.WorkflowFailed, argo.WorkflowError:
			pipeline.Status.SynchronizationState = pipelinesv1.Failed
		case argo.WorkflowSucceeded:
			pipeline.Status.SynchronizationState = pipelinesv1.Succeeded
			pipeline.Status.Id = string(*workflow.Status.Nodes[workflow.Name].Outputs.Parameters[0].Value)
		}

		if err := r.Status().Update(ctx, &pipeline); err != nil {
			return err
		}
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
