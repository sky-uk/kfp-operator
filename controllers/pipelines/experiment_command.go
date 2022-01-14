package pipelines

import (
	"context"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

type ExperimentCommand interface {
	execute(*ExperimentReconciler, context.Context, *pipelinesv1.Experiment) error
}

type SetExperimentStatus struct {
	Status pipelinesv1.Status
}

func SetExperimentSynchronizationStateOnly(experiment *pipelinesv1.Experiment, state pipelinesv1.SynchronizationState) SetExperimentStatus {
	return SetExperimentStatus{
		Status: pipelinesv1.Status{
			KfpId:                experiment.Status.KfpId,
			Version:              experiment.Status.Version,
			SynchronizationState: state,
		},
	}
}

func (srcs SetExperimentStatus) execute(reconciler *ExperimentReconciler, ctx context.Context, rc *pipelinesv1.Experiment) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info("setting experiment status", LogKeys.OldStatus, rc.Status, LogKeys.NewStatus, srcs.Status)

	rc.Status = srcs.Status

	return reconciler.Client.Status().Update(ctx, rc)
}

type CreateExperimentWorkflow struct {
	Workflow argo.Workflow
}

func (cw CreateExperimentWorkflow) execute(reconciler *ExperimentReconciler, ctx context.Context, rc *pipelinesv1.Experiment) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info("creating child workflow", LogKeys.Workflow, cw.Workflow)

	if err := ctrl.SetControllerReference(rc, &cw.Workflow, reconciler.Scheme); err != nil {
		return err
	}

	if err := reconciler.Client.Create(ctx, &cw.Workflow); err != nil {
		return err
	}

	return nil
}

type DeleteExperimentWorkflows struct {
	Workflows []argo.Workflow
}

func (dw DeleteExperimentWorkflows) execute(reconciler *ExperimentReconciler, ctx context.Context, _ *pipelinesv1.Experiment) error {
	logger := log.FromContext(ctx)

	for i := range dw.Workflows {
		workflow := &dw.Workflows[i]
		workflowDebugOptions := pipelinesv1.DebugOptionsFromAnnotations(ctx, workflow.ObjectMeta.Annotations)
		if !workflowDebugOptions.KeepWorkflows {
			logger.V(1).Info("deleting child workflow", LogKeys.Workflow, workflow)
			if err := reconciler.Client.Delete(ctx, workflow); err != nil {
				return err
			}
		} else {
			logger.V(2).Info("keeping child workflow", LogKeys.Workflow, workflow)
		}
	}

	return nil
}

type AcquireExperiment struct {
}

func (ar AcquireExperiment) execute(reconciler *ExperimentReconciler, ctx context.Context, rc *pipelinesv1.Experiment) error {
	logger := log.FromContext(ctx)

	if !containsString(rc.ObjectMeta.Finalizers, finalizerName) {
		logger.V(2).Info("adding finalizer")
		rc.ObjectMeta.Finalizers = append(rc.ObjectMeta.Finalizers, finalizerName)
		return reconciler.Client.Update(ctx, rc)
	}

	return nil
}

type ReleaseExperiment struct {
}

func (rr ReleaseExperiment) execute(reconciler *ExperimentReconciler, ctx context.Context, rc *pipelinesv1.Experiment) error {
	logger := log.FromContext(ctx)

	if containsString(rc.ObjectMeta.Finalizers, finalizerName) {
		logger.V(2).Info("removing finalizer")
		rc.ObjectMeta.Finalizers = removeString(rc.ObjectMeta.Finalizers, finalizerName)
		return reconciler.Client.Update(ctx, rc)
	}

	return nil
}
