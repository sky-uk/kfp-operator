package pipelines

import (
	"context"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

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
