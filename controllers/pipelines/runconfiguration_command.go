package pipelines

import (
	"context"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

type RunConfigurationCommand interface {
	execute(*RunConfigurationReconciler, context.Context, *pipelinesv1.RunConfiguration) error
}

type SetRunConfigurationStatus struct {
	Status pipelinesv1.Status
}

func (srcs SetRunConfigurationStatus) execute(reconciler *RunConfigurationReconciler, ctx context.Context, rc *pipelinesv1.RunConfiguration) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info("setting run configuration status", LogKeys.OldStatus, rc.Status, LogKeys.NewStatus, srcs.Status)

	rc.Status = srcs.Status

	return reconciler.Status().Update(ctx, rc)
}

type CreateRunConfigurationWorkflow struct {
	Workflow argo.Workflow
}

func (cw CreateRunConfigurationWorkflow) execute(reconciler *RunConfigurationReconciler, ctx context.Context, rc *pipelinesv1.RunConfiguration) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info("creating child workflow", LogKeys.Workflow, cw.Workflow)

	if err := ctrl.SetControllerReference(rc, &cw.Workflow, reconciler.Scheme); err != nil {
		return err
	}

	if err := reconciler.Create(ctx, &cw.Workflow); err != nil {
		return err
	}

	return nil
}

type DeleteRunConfigurationWorkflows struct {
	Workflows []argo.Workflow
}

func (dw DeleteRunConfigurationWorkflows) execute(reconciler *RunConfigurationReconciler, ctx context.Context, _ *pipelinesv1.RunConfiguration) error {
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
			logger.V(2).Info("keeping child workflow", LogKeys.Workflow, workflow)
		}
	}

	return nil
}

type AcquireRunConfiguration struct {
}

func (ar AcquireRunConfiguration) execute(reconciler *RunConfigurationReconciler, ctx context.Context, rc *pipelinesv1.RunConfiguration) error {
	logger := log.FromContext(ctx)

	if !containsString(rc.ObjectMeta.Finalizers, finalizerName) {
		logger.V(2).Info("adding finalizer")
		rc.ObjectMeta.Finalizers = append(rc.ObjectMeta.Finalizers, finalizerName)
		return reconciler.Update(ctx, rc)
	}

	return nil
}

type ReleaseRunConfiguration struct {
}

func (rr ReleaseRunConfiguration) execute(reconciler *RunConfigurationReconciler, ctx context.Context, rc *pipelinesv1.RunConfiguration) error {
	logger := log.FromContext(ctx)

	if containsString(rc.ObjectMeta.Finalizers, finalizerName) {
		logger.V(2).Info("removing finalizer")
		rc.ObjectMeta.Finalizers = removeString(rc.ObjectMeta.Finalizers, finalizerName)
		return reconciler.Update(ctx, rc)
	}

	return nil
}
