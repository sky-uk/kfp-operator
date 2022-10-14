package pipelines

import (
	"context"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PipelineStateHandler struct {
	WorkflowFactory    WorkflowFactory[*pipelinesv1.Pipeline]
	WorkflowRepository WorkflowRepository
}

func (st PipelineStateHandler) stateTransition(ctx context.Context, pipeline *pipelinesv1.Pipeline) (commands []Command) {
	switch pipeline.Status.SynchronizationState {
	case apis.Creating:
		commands = st.onCreating(ctx, pipeline,
			st.WorkflowRepository.GetByLabels(ctx, pipeline.GetNamespace(),
				CommonWorkflowLabels(pipeline, WorkflowConstants.CreateOperationLabel)))
	case apis.Succeeded, apis.Failed:
		if !pipeline.ObjectMeta.DeletionTimestamp.IsZero() {
			commands = st.onDelete(ctx, pipeline)
		} else {
			commands = st.onSucceededOrFailed(ctx, pipeline)
		}
	case apis.Updating:
		commands = st.onUpdating(ctx, pipeline,
			st.WorkflowRepository.GetByLabels(ctx, pipeline.GetNamespace(),
				CommonWorkflowLabels(pipeline, WorkflowConstants.UpdateOperationLabel)))
	case apis.Deleting:
		commands = st.onDeleting(ctx, pipeline,
			st.WorkflowRepository.GetByLabels(ctx, pipeline.GetNamespace(),
				CommonWorkflowLabels(pipeline, WorkflowConstants.DeleteOperationLabel)))
	case apis.Deleted:
	default:
		commands = st.onUnknown(ctx, pipeline)
	}

	if pipeline.Status.SynchronizationState == apis.Deleted {
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
					WithSynchronizationState(apis.Failed).
					WithVersion(newPipelineVersion).
					WithMessage(WorkflowConstants.ConstructionFailedError),
			}
		}

		return []Command{
			*From(pipeline.Status).
				WithSynchronizationState(apis.Updating).
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
				WithSynchronizationState(apis.Failed).
				WithVersion(newPipelineVersion).
				WithMessage(WorkflowConstants.ConstructionFailedError),
		}
	}

	return []Command{
		*From(pipeline.Status).
			WithSynchronizationState(apis.Creating).
			WithVersion(newPipelineVersion),
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st PipelineStateHandler) onDelete(ctx context.Context, pipeline *pipelinesv1.Pipeline) []Command {
	logger := log.FromContext(ctx)
	logger.Info("deletion requested, deleting")

	if pipeline.Status.KfpId == "" {
		return []Command{
			*From(pipeline.Status).WithSynchronizationState(apis.Deleted),
		}
	}

	workflow, _ := st.WorkflowFactory.ConstructDeletionWorkflow(pipeline)

	return []Command{
		*From(pipeline.Status).WithSynchronizationState(apis.Deleting),
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
	var targetState apis.SynchronizationState

	if pipeline.Status.KfpId == "" {
		logger.Info("no kfpId exists, creating")
		workflow, err = st.WorkflowFactory.ConstructCreationWorkflow(pipeline)
		targetState = apis.Creating
	} else {
		logger.Info("kfpId exists, updating")
		workflow, err = st.WorkflowFactory.ConstructUpdateWorkflow(pipeline)
		targetState = apis.Updating
	}

	if err != nil {
		failureMessage := WorkflowConstants.ConstructionFailedError
		logger.Info(fmt.Sprintf("%s, failing pipeline", failureMessage))
		return []Command{
			*From(pipeline.Status).
				WithSynchronizationState(apis.Failed).
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
			*From(pipeline.Status).WithSynchronizationState(apis.Failed).WithMessage(failureMessage),
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
		setStatusCommand = From(pipeline.Status).WithSynchronizationState(apis.Succeeded)
	} else {
		var failureMessage string
		if failed != nil {
			failureMessage = "pipeline update failed"
		} else {
			failureMessage = "pipeline updating progress unknown"
		}

		logger.Info(fmt.Sprintf("%s, failing pipeline", failureMessage))
		setStatusCommand = From(pipeline.Status).
			WithSynchronizationState(apis.Failed).
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
		setStatusCommand = From(pipeline.Status).WithSynchronizationState(apis.Deleted)
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
				WithSynchronizationState(apis.Failed).
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
			return From(pipeline.Status).WithSynchronizationState(apis.Failed).WithMessage(failureMessage)
		}

		result, err := getWorkflowOutput(succeeded, WorkflowConstants.ProviderOutputParameterName)

		if err != nil {
			failureMessage := "could not retrieve kfpId"
			logger.Error(err, fmt.Sprintf("%s, failing pipeline", failureMessage))
			return From(pipeline.Status).WithSynchronizationState(apis.Failed)
		}

		if result.ProviderError != "" {
			failureMessage := "pipeline creation succeeded but error returned"
			logger.Info(failureMessage)
			return From(pipeline.Status).
				WithSynchronizationState(apis.Failed).
				WithKfpId(result.Id).
				WithMessage(failureMessage)
		}

		logger.Info("pipeline creation succeeded")
		return From(pipeline.Status).WithSynchronizationState(apis.Succeeded).WithKfpId(result.Id)
	}

	return []Command{
		*statusAfterCreating(),
		DeleteWorkflows{
			Workflows: creationWorkflows,
		},
	}
}
