package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PipelineCommand interface {
	execute(*PipelineReconciler, context.Context, *pipelinesv1.Pipeline) error
}

type SetPipelineStatus struct {
	Status pipelinesv1.Status
}

func (sps SetPipelineStatus) execute(reconciler *PipelineReconciler, ctx context.Context, pipeline *pipelinesv1.Pipeline) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info("setting pipeline status", LogKeys.OldStatus, pipeline.Status, LogKeys.NewStatus, sps.Status)

	pipeline.Status = sps.Status

	return reconciler.Status().Update(ctx, pipeline)
}

type CreatePipelineWorkflow struct {
	Workflow argo.Workflow
}

func (cw CreatePipelineWorkflow) execute(reconciler *PipelineReconciler, ctx context.Context, pipeline *pipelinesv1.Pipeline) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info("creating child workflow", LogKeys.Workflow, cw.Workflow)

	if err := ctrl.SetControllerReference(pipeline, &cw.Workflow, reconciler.Scheme); err != nil {
		return err
	}

	if err := reconciler.Create(ctx, &cw.Workflow); err != nil {
		return err
	}

	return nil
}

type DeletePipelineWorkflows struct {
	Workflows []argo.Workflow
}

func (dw DeletePipelineWorkflows) execute(reconciler *PipelineReconciler, ctx context.Context, _ *pipelinesv1.Pipeline) error {
	logger := log.FromContext(ctx)

	for i := range dw.Workflows {
		workflow := &dw.Workflows[i]
		workflowDebugOptions := pipelinesv1.DebugOptionsFromAnnotations(ctx, workflow.ObjectMeta.Annotations)
		if !workflowDebugOptions.KeepWorkflows {
			logger.V(1).Info("deleting child workflow", LogKeys.Workflow, workflow)
			if err := reconciler.Delete(ctx, workflow); err != nil {
				return err
			}
		} else {
			logger.V(1).Info("keeping child workflow", LogKeys.Workflow, workflow)
		}
	}

	return nil
}

type DeletePipeline struct {
}

func (dp DeletePipeline) execute(reconciler *PipelineReconciler, ctx context.Context, pipeline *pipelinesv1.Pipeline) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info("releasing pipeline resource")

	return reconciler.RemoveFinalizer(ctx, pipeline)
}
