package pipelines

import (
	"context"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

type PipelineCommand interface {
	execute(*PipelineReconciler, context.Context, *pipelinesv1.Pipeline) error
}

type SetPipelineStatus struct {
	Status pipelinesv1.Status
}

func (sps SetPipelineStatus) execute(reconciler *PipelineReconciler, ctx context.Context, pipeline *pipelinesv1.Pipeline) error {
	pipeline.Status = sps.Status

	return reconciler.Status().Update(ctx, pipeline)
}

type CreatePipelineWorkflow struct {
	Workflow argo.Workflow
}

func (cw CreatePipelineWorkflow) execute(reconciler *PipelineReconciler, ctx context.Context, pipeline *pipelinesv1.Pipeline) error {
	return reconciler.CreateChildWorkflow(ctx, pipeline, cw.Workflow)
}

type DeletePipelineWorkflows struct {
	Workflows []argo.Workflow
}

func (dw DeletePipelineWorkflows) execute(reconciler *PipelineReconciler, ctx context.Context, _ *pipelinesv1.Pipeline) error {
	for i := range dw.Workflows {
		workflow := &dw.Workflows[i]
		workflowDebugOptions := pipelinesv1.DebugOptionsFromAnnotations(ctx, workflow.ObjectMeta.Annotations)
		if !workflowDebugOptions.KeepWorkflows {
			if err := reconciler.Delete(ctx, workflow); err != nil {
				return err
			}
		}
	}

	return nil
}

type DeletePipeline struct {
}

func (dp DeletePipeline) execute(reconciler *PipelineReconciler, ctx context.Context, pipeline *pipelinesv1.Pipeline) error {
	return reconciler.RemoveFinalizer(ctx, pipeline)
}
