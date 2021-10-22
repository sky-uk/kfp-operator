package pipelines

import (
	"context"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

type RunConfigurationCommand interface {
	execute(*RunConfigurationReconciler, context.Context, *pipelinesv1.RunConfiguration) error
}

type SetRunConfigurationStatus struct {
	Status pipelinesv1.Status
}

func (sps SetRunConfigurationStatus) execute(reconciler *RunConfigurationReconciler, ctx context.Context, rc *pipelinesv1.RunConfiguration) error {
	rc.Status = sps.Status

	return reconciler.Status().Update(ctx, rc)
}

type CreateRunConfigurationWorkflow struct {
	Workflow argo.Workflow
}

func (cw CreateRunConfigurationWorkflow) execute(reconciler *RunConfigurationReconciler, ctx context.Context, rc *pipelinesv1.RunConfiguration) error {
	return reconciler.CreateChildWorkflow(ctx, rc, cw.Workflow)
}

type DeleteRunConfigurationWorkflows struct {
	Workflows []argo.Workflow
}

func (dw DeleteRunConfigurationWorkflows) execute(reconciler *RunConfigurationReconciler, ctx context.Context, _ *pipelinesv1.RunConfiguration) error {
	for i := range dw.Workflows {
		if err := reconciler.Delete(ctx, &dw.Workflows[i]); err != nil {
			return err
		}
	}

	return nil
}

type DeleteRunConfiguration struct {
}

func (dp DeleteRunConfiguration) execute(reconciler *RunConfigurationReconciler, ctx context.Context, rc *pipelinesv1.RunConfiguration) error {
	return reconciler.RemoveFinalizer(ctx, *rc)
}
