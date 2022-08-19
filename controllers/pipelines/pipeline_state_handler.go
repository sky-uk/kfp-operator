package pipelines

import (
	"context"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PipelineStateHandler struct {
	WorkflowFactory    WorkflowFactory[*pipelinesv1.Pipeline]
	WorkflowRepository WorkflowRepository
}

func (st PipelineStateHandler) stateTransition(ctx context.Context, pipeline *pipelinesv1.Pipeline) (commands []Command) {
	switch pipeline.Status.SynchronizationState {
	case pipelinesv1.Creating:
		commands = st.onCreating(ctx, pipeline,
			st.WorkflowRepository.GetByLabels(ctx, pipeline.GetNamespace(),
				CommonWorkflowLabels(pipeline, WorkflowConstants.CreateOperationLabel)))
	case pipelinesv1.Succeeded, pipelinesv1.Failed:
		if !pipeline.ObjectMeta.DeletionTimestamp.IsZero() {
			commands = st.onDelete(ctx, pipeline)
		} else {
			commands = st.onSucceededOrFailed(ctx, pipeline)
		}
	case pipelinesv1.Updating:
		commands = st.onUpdating(ctx, pipeline,
			st.WorkflowRepository.GetByLabels(ctx, pipeline.GetNamespace(),
				CommonWorkflowLabels(pipeline, WorkflowConstants.UpdateOperationLabel)))
	case pipelinesv1.Deleting:
		commands = st.onDeleting(ctx, pipeline,
			st.WorkflowRepository.GetByLabels(ctx, pipeline.GetNamespace(),
				CommonWorkflowLabels(pipeline, WorkflowConstants.DeleteOperationLabel)))
	case pipelinesv1.Deleted:
	default:
		commands = st.onUnknown(ctx, pipeline)
	}

	if pipeline.Status.SynchronizationState == pipelinesv1.Deleted {
		commands = append([]Command{ReleaseResource{}}, commands...)
	} else {
		commands = append([]Command{AcquireResource{}}, commands...)
	}

	return
}

func (st *PipelineStateHandler) StateTransition(ctx context.Context, pipeline *pipelinesv1.Pipeline) []Command {
	logger := log.FromContext(ctx)
	logger.Info("state transition start")

	stateTransitionCommands := st.stateTransition(ctx, pipeline)
	return alwaysSetObservedGeneration(ctx, stateTransitionCommands, pipeline)
}

func (st PipelineStateHandler) onUnknown(ctx context.Context, pipeline *pipelinesv1.Pipeline) []Command {
	logger := log.FromContext(ctx)

	newPipelineVersion := pipeline.Spec.ComputeVersion()

	if pipeline.Status.KfpId != "" {
		logger.Info("empty state but kfpId already exists, updating pipeline")
		workflow, err := st.WorkflowFactory.ConstructUpdateWorkflow(pipeline)

		if err != nil {
			logger.Error(err, fmt.Sprintf("%s, failing pipeline", WorkflowConstants.ConstructionFailedError))

			return []Command{
				*From(pipeline.Status).
					WithSynchronizationState(pipelinesv1.Failed).
					WithVersion(newPipelineVersion).
					WithMessage(WorkflowConstants.ConstructionFailedError),
			}
		}

		return []Command{
			*From(pipeline.Status).
				WithSynchronizationState(pipelinesv1.Updating).
				WithVersion(newPipelineVersion),
			CreateWorkflow{Workflow: *workflow},
		}
	}

	logger.Info("empty state, creating pipeline")
	workflow, err := st.WorkflowFactory.ConstructCreationWorkflow(pipeline)

	if err != nil {
		logger.Error(err, fmt.Sprintf("%s, failing pipeline", WorkflowConstants.ConstructionFailedError))

		return []Command{
			*From(pipeline.Status).
				WithSynchronizationState(pipelinesv1.Failed).
				WithVersion(newPipelineVersion).
				WithMessage(WorkflowConstants.ConstructionFailedError),
		}
	}

	return []Command{
		*From(pipeline.Status).
			WithSynchronizationState(pipelinesv1.Creating).
			WithVersion(newPipelineVersion),
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st PipelineStateHandler) onDelete(ctx context.Context, pipeline *pipelinesv1.Pipeline) []Command {
	logger := log.FromContext(ctx)
	logger.Info("deletion requested, deleting")

	if pipeline.Status.KfpId == "" {
		return []Command{
			*From(pipeline.Status).WithSynchronizationState(pipelinesv1.Deleted),
		}
	}

	workflow, _ := st.WorkflowFactory.ConstructDeletionWorkflow(pipeline)

	return []Command{
		*From(pipeline.Status).WithSynchronizationState(pipelinesv1.Deleting),
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st PipelineStateHandler) onSucceededOrFailed(ctx context.Context, pipeline *pipelinesv1.Pipeline) []Command {
	logger := log.FromContext(ctx)
	newPipelineVersion := pipeline.Spec.ComputeVersion()

	if pipeline.Status.Version == newPipelineVersion {
		logger.V(2).Info("pipeline version has not changed")
		return []Command{}
	}

	var workflow *argo.Workflow
	var err error
	var targetState pipelinesv1.SynchronizationState

	if pipeline.Status.KfpId == "" {
		logger.Info("no kfpId exists, creating")
		workflow, err = st.WorkflowFactory.ConstructCreationWorkflow(pipeline)
		targetState = pipelinesv1.Creating
	} else {
		logger.Info("kfpId exists, updating")
		workflow, err = st.WorkflowFactory.ConstructUpdateWorkflow(pipeline)
		targetState = pipelinesv1.Updating
	}

	if err != nil {
		failureMessage := WorkflowConstants.ConstructionFailedError
		logger.Info(fmt.Sprintf("%s, failing pipeline", failureMessage))
		return []Command{
			*From(pipeline.Status).
				WithSynchronizationState(pipelinesv1.Failed).
				WithVersion(newPipelineVersion).
				WithMessage(failureMessage),
		}
	}

	return []Command{
		*From(pipeline.Status).
			WithSynchronizationState(targetState).
			WithVersion(newPipelineVersion),
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st PipelineStateHandler) onUpdating(ctx context.Context, pipeline *pipelinesv1.Pipeline, updateWorkflows []argo.Workflow) []Command {
	logger := log.FromContext(ctx)

	if pipeline.Status.Version == "" || pipeline.Status.KfpId == "" {
		failureMessage := "updating pipeline with empty version or kfpId"
		logger.Info(fmt.Sprintf("%s, failing pipeline", failureMessage))

		return []Command{
			*From(pipeline.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage),
		}
	}

	inProgress, succeeded, failed := latestWorkflowByPhase(updateWorkflows)

	if inProgress != nil {
		logger.V(2).Info("pipeline update in progress")
		return []Command{}
	}

	var setStatusCommand *SetStatus

	if succeeded != nil {
		logger.Info("pipeline update succeeded")
		setStatusCommand = From(pipeline.Status).WithSynchronizationState(pipelinesv1.Succeeded)
	} else {
		var failureMessage string
		if failed != nil {
			failureMessage = "pipeline update failed"
		} else {
			failureMessage = "pipeline updating progress unknown"
		}

		logger.Info(fmt.Sprintf("%s, failing pipeline", failureMessage))
		setStatusCommand = From(pipeline.Status).
			WithSynchronizationState(pipelinesv1.Failed).
			WithMessage(failureMessage)
	}

	return []Command{
		*setStatusCommand,
		DeleteWorkflows{
			Workflows: updateWorkflows,
		},
	}
}

func (st PipelineStateHandler) onDeleting(ctx context.Context, pipeline *pipelinesv1.Pipeline, deletionWorkflows []argo.Workflow) []Command {
	logger := log.FromContext(ctx)

	inProgress, succeeded, failed := latestWorkflowByPhase(deletionWorkflows)

	if inProgress != nil {
		logger.V(2).Info("pipeline deletion in progress")
		return []Command{}
	}

	var setStatusCommand *SetStatus

	if succeeded != nil {
		logger.Info("pipeline deletion succeeded")
		setStatusCommand = From(pipeline.Status).WithSynchronizationState(pipelinesv1.Deleted)
	} else {
		var failureMessage string

		if failed != nil {
			failureMessage = "pipeline deletion failed"
		} else {
			failureMessage = "pipeline deletion progress unknown"
		}

		logger.Info(fmt.Sprintf("%s, failing pipeline", failureMessage))
		setStatusCommand = From(pipeline.Status).WithMessage(failureMessage)
	}

	return []Command{
		*setStatusCommand,
		DeleteWorkflows{
			Workflows: deletionWorkflows,
		},
	}
}

func (st PipelineStateHandler) onCreating(ctx context.Context, pipeline *pipelinesv1.Pipeline, creationWorkflows []argo.Workflow) []Command {
	logger := log.FromContext(ctx)

	if pipeline.Status.Version == "" {
		failureMessage := "creating pipeline with empty version"
		logger.Info(fmt.Sprintf("%s, failing pipeline", failureMessage))

		return []Command{
			*From(pipeline.Status).
				WithSynchronizationState(pipelinesv1.Failed).
				WithMessage(failureMessage),
		}
	}

	inProgress, succeeded, failed := latestWorkflowByPhase(creationWorkflows)

	if inProgress != nil {
		logger.V(2).Info("pipeline creation in progress")
		return []Command{}
	}

	statusAfterCreating := func() *SetStatus {
		if succeeded == nil {
			var failureMessage string
			if failed != nil {
				failureMessage = "pipeline creation failed"
			} else {
				failureMessage = "pipeline creation progress unknown"
			}

			logger.Info(fmt.Sprintf("%s, failing pipeline", failureMessage))
			return From(pipeline.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage)
		}

		idResult, err := getWorkflowOutput(succeeded, PipelineWorkflowConstants.PipelineIdParameterName)

		if err != nil {
			failureMessage := "could not retrieve kfpId"
			logger.Error(err, fmt.Sprintf("%s, failing pipeline", failureMessage))
			return From(pipeline.Status).WithSynchronizationState(pipelinesv1.Failed)
		}

		versionResult, _ := getWorkflowOutput(succeeded, PipelineWorkflowConstants.PipelineVersionParameterName)

		if versionResult == "" {
			failureMessage := "pipeline creation succeeded but version upload failed"
			logger.Info(failureMessage)
			return From(pipeline.Status).
				WithSynchronizationState(pipelinesv1.Failed).
				WithKfpId(idResult).
				WithMessage(failureMessage)
		}

		logger.Info("pipeline creation succeeded")
		return From(pipeline.Status).WithSynchronizationState(pipelinesv1.Succeeded).WithKfpId(idResult)
	}

	return []Command{
		*statusAfterCreating(),
		DeleteWorkflows{
			Workflows: creationWorkflows,
		},
	}
}
