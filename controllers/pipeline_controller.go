package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

	var err error
	switch pipeline.Status.SynchronizationState {
	case pipelinesv1.Unknown:
		err = r.onUnknown(ctx, pipeline)
	case pipelinesv1.Creating:
		err = r.onCreating(ctx, pipeline)
	case pipelinesv1.Succeeded:
		err = r.onSucceeded(ctx, pipeline)
	case pipelinesv1.Updating:
		err = r.onUpdating(ctx, pipeline)
	}

	if err != nil {
		logger.Error(err, fmt.Sprintf("unable to transition from state %s", pipeline.Status.SynchronizationState))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *PipelineReconciler) onUnknown(ctx context.Context, pipeline pipelinesv1.Pipeline) error {
	workflow, err := constructCreationWorkflow(&pipeline)

	if err != nil {
		return err
	}

	r.createChildWorkflow(ctx, &pipeline, workflow)

	pipelineVersion, err := pipelinesv1.ComputeVersion(pipeline.Spec)
	if err != nil {
		return err
	}

	pipeline.Status.Version = pipelineVersion
	pipeline.Status.SynchronizationState = pipelinesv1.Creating

	if err := r.Status().Update(ctx, &pipeline); err != nil {
		return err
	}

	return nil
}

func (r *PipelineReconciler) onSucceeded(ctx context.Context, pipeline pipelinesv1.Pipeline) error {
	newPipelineVersion, err := pipelinesv1.ComputeVersion(pipeline.Spec)

	if err != nil {
		return err
	}

	if pipeline.Status.Version == newPipelineVersion {
		return nil
	}

	workflow, err := constructUpdateWorkflow(&pipeline)

	if err != nil {
		return err
	}

	r.createChildWorkflow(ctx, &pipeline, workflow)

	pipeline.Status.Version = newPipelineVersion
	pipeline.Status.SynchronizationState = pipelinesv1.Updating

	if err := r.Status().Update(ctx, &pipeline); err != nil {
		return err
	}

	return nil
}

func (r *PipelineReconciler) onUpdating(ctx context.Context, pipeline pipelinesv1.Pipeline) error {
	childWorkflow, err := r.getChildWorkflow(ctx, pipeline, "update-pipeline")

	if err != nil {
		return err
	}

	if childWorkflow != nil {
		switch childWorkflow.Status.Phase {
		case argo.WorkflowFailed, argo.WorkflowError:
			pipeline.Status.SynchronizationState = pipelinesv1.Failed
		case argo.WorkflowSucceeded:
			pipeline.Status.SynchronizationState = pipelinesv1.Succeeded
		}

		if err := r.Status().Update(ctx, &pipeline); err != nil {
			return err
		}
	}

	return nil
}

func (r *PipelineReconciler) onCreating(ctx context.Context, pipeline pipelinesv1.Pipeline) error {
	childWorkflow, err := r.getChildWorkflow(ctx, pipeline, "create-pipeline")

	if err != nil {
		return err
	}

	if childWorkflow != nil {
		switch childWorkflow.Status.Phase {
		case argo.WorkflowFailed, argo.WorkflowError:
			pipeline.Status.SynchronizationState = pipelinesv1.Failed
		case argo.WorkflowSucceeded:
			pipeline.Status.SynchronizationState = pipelinesv1.Succeeded
			pipeline.Status.Id = string(*childWorkflow.Status.Nodes[childWorkflow.Name].Outputs.Parameters[0].Value)
		}

		if err := r.Status().Update(ctx, &pipeline); err != nil {
			return err
		}
	}

	return nil
}

func (r *PipelineReconciler) createChildWorkflow(ctx context.Context, pipeline *pipelinesv1.Pipeline, workflow *argo.Workflow) error {
	if err := ctrl.SetControllerReference(pipeline, workflow, r.Scheme); err != nil {
		return err
	}

	if err := r.Create(ctx, workflow); err != nil {
		return err
	}

	return nil
}

func (r *PipelineReconciler) getChildWorkflow(ctx context.Context, pipeline pipelinesv1.Pipeline, operation string) (*argo.Workflow, error) {
	workflow := argo.Workflow{}

	name := types.NamespacedName{Name: operation + "-" + pipeline.ObjectMeta.Name, Namespace: pipeline.ObjectMeta.Namespace}
	err := r.Get(ctx, name, &workflow)

	if err != nil {
		return nil, err
	}

	return &workflow, nil
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
