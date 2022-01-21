package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PipelineStateHandler struct {
	WorkflowFactory    PipelineWorkflowFactory
	WorkflowRepository WorkflowRepository
}

func (st PipelineStateHandler) StateTransition(ctx context.Context, pipeline *pipelinesv1.Pipeline) (commands []PipelineCommand) {
	logger := log.FromContext(ctx)
	logger.Info("state transition start")

	switch pipeline.Status.SynchronizationState {
	case pipelinesv1.Creating:
		commands = st.onCreating(ctx, pipeline,
			st.WorkflowRepository.GetByOperation(ctx,
				PipelineWorkflowConstants.CreateOperationLabel,
				pipeline.NamespacedName(),
				PipelineWorkflowConstants.PipelineNameLabelKey))
	case pipelinesv1.Succeeded, pipelinesv1.Failed:
		if !pipeline.ObjectMeta.DeletionTimestamp.IsZero() {
			commands = st.onDelete(ctx, pipeline)
		} else {
			commands = st.onSucceededOrFailed(ctx, pipeline)
		}
	case pipelinesv1.Updating:
		commands = st.onUpdating(ctx, pipeline,
			st.WorkflowRepository.GetByOperation(ctx,
				PipelineWorkflowConstants.UpdateOperationLabel,
				pipeline.NamespacedName(),
				PipelineWorkflowConstants.PipelineNameLabelKey))
	case pipelinesv1.Deleting:
		commands = st.onDeleting(ctx, pipeline,
			st.WorkflowRepository.GetByOperation(ctx,
				PipelineWorkflowConstants.DeleteOperationLabel,
				pipeline.NamespacedName(),
				PipelineWorkflowConstants.PipelineNameLabelKey))
	case pipelinesv1.Deleted:
	default:
		commands = st.onUnknown(ctx, pipeline)
	}

	if pipeline.Status.SynchronizationState == pipelinesv1.Deleted {
		commands = append([]PipelineCommand{ReleasePipeline{}}, commands...)
	} else {
		commands = append([]PipelineCommand{AcquirePipeline{}}, commands...)
	}

	return
}

func (st PipelineStateHandler) onUnknown(ctx context.Context, pipeline *pipelinesv1.Pipeline) []PipelineCommand {
	logger := log.FromContext(ctx)

	newPipelineVersion := pipeline.Spec.ComputeVersion()

	if pipeline.Status.KfpId != "" {
		logger.Info("empty state but kfpId already exists, updating pipeline")
		workflow, err := st.WorkflowFactory.ConstructUpdateWorkflow(ctx, pipeline)

		if err != nil {
			failureMessage := "error constructing update workflow, failing pipeline"
			logger.Error(err, failureMessage)

			return []PipelineCommand{
				*FromState(pipeline).
					To(pipelinesv1.Failed).
					WithMessage(failureMessage),
			}
		}

		return []PipelineCommand{
			*FromState(pipeline).
				To(pipelinesv1.Updating).
				WithVersion(newPipelineVersion),
			CreatePipelineWorkflow{Workflow: *workflow},
		}
	}

	logger.Info("empty state, creating pipeline")
	workflow, err := st.WorkflowFactory.ConstructCreationWorkflow(ctx, pipeline)

	if err != nil {
		failureMessage := "error constructing creation workflow, failing pipeline"
		logger.Error(err, failureMessage)

		return []PipelineCommand{
			*FromState(pipeline).
				To(pipelinesv1.Failed).
				WithVersion(newPipelineVersion).
				WithMessage(failureMessage),
		}
	}

	return []PipelineCommand{
		*FromState(pipeline).
			To(pipelinesv1.Creating).
			WithVersion(newPipelineVersion),
		CreatePipelineWorkflow{Workflow: *workflow},
	}
}

func (st PipelineStateHandler) onDelete(ctx context.Context, pipeline *pipelinesv1.Pipeline) []PipelineCommand {
	logger := log.FromContext(ctx)
	logger.Info("deletion requested, deleting")

	if pipeline.Status.KfpId == "" {
		return []PipelineCommand{
			*FromState(pipeline).To(pipelinesv1.Deleted),
		}
	}

	workflow, err := st.WorkflowFactory.ConstructDeletionWorkflow(ctx, pipeline)

	if err != nil {
		failureMessage := "error constructing deletion workflow, failing pipeline"
		logger.Error(err, failureMessage)

		return []PipelineCommand{
			*FromState(pipeline).
				To(pipelinesv1.Failed).
				WithMessage(failureMessage),
		}
	}

	return []PipelineCommand{
		*FromState(pipeline).To(pipelinesv1.Deleting),
		CreatePipelineWorkflow{Workflow: *workflow},
	}
}

func (st PipelineStateHandler) onSucceededOrFailed(ctx context.Context, pipeline *pipelinesv1.Pipeline) []PipelineCommand {
	logger := log.FromContext(ctx)
	newPipelineVersion := pipeline.Spec.ComputeVersion()

	if pipeline.Status.Version == newPipelineVersion {
		logger.V(2).Info("pipeline version has not changed")
		return []PipelineCommand{}
	}

	var workflow *argo.Workflow
	var err error
	var targetState pipelinesv1.SynchronizationState

	if pipeline.Status.KfpId == "" {
		logger.Info("no kfpId exists, creating")
		workflow, err = st.WorkflowFactory.ConstructCreationWorkflow(ctx, pipeline)
		targetState = pipelinesv1.Creating
	} else {
		logger.Info("kfpId exists, updating")
		workflow, err = st.WorkflowFactory.ConstructUpdateWorkflow(ctx, pipeline)
		targetState = pipelinesv1.Updating
	}

	if err != nil {
		logger.Info("error constructing workflow, failing pipeline")
		return []PipelineCommand{
			*FromState(pipeline).
				To(pipelinesv1.Failed).
				WithVersion(newPipelineVersion),
		}
	}

	return []PipelineCommand{
		*FromState(pipeline).
			To(targetState).
			WithVersion(newPipelineVersion),
		CreatePipelineWorkflow{Workflow: *workflow},
	}
}

func (st PipelineStateHandler) onUpdating(ctx context.Context, pipeline *pipelinesv1.Pipeline, updateWorkflows []argo.Workflow) []PipelineCommand {
	logger := log.FromContext(ctx)

	if pipeline.Status.Version == "" || pipeline.Status.KfpId == "" {
		failureMessage := "updating pipeline with empty version or kfpId, failing pipeline"
		logger.Info(failureMessage)

		return []PipelineCommand{
			*FromState(pipeline).To(pipelinesv1.Failed).WithMessage(failureMessage),
		}
	}

	inProgress, succeeded, failed := latestWorkflowByPhase(updateWorkflows)

	if inProgress != nil {
		logger.V(2).Info("pipeline update in progress")
		return []PipelineCommand{}
	}

	var newState pipelinesv1.SynchronizationState

	if succeeded != nil {
		logger.Info("pipeline update succeeded")
		newState = pipelinesv1.Succeeded
	} else {
		if failed != nil {
			logger.Info("pipeline update failed")
		} else {
			logger.Info("pipeline updating progress unknown, failing pipeline")
		}
		newState = pipelinesv1.Failed
	}

	return []PipelineCommand{
		*FromState(pipeline).To(newState),
		DeletePipelineWorkflows{
			Workflows: updateWorkflows,
		},
	}
}

func (st PipelineStateHandler) onDeleting(ctx context.Context, pipeline *pipelinesv1.Pipeline, deletionWorkflows []argo.Workflow) []PipelineCommand {
	logger := log.FromContext(ctx)

	inProgress, succeeded, failed := latestWorkflowByPhase(deletionWorkflows)

	if inProgress != nil {
		logger.V(2).Info("pipeline deletion in progress")
		return []PipelineCommand{}
	}

	var statusCommand *SetPipelineStatus

	if succeeded != nil {
		logger.Info("pipeline deletion succeeded")
		statusCommand = FromState(pipeline).To(pipelinesv1.Deleted)
	} else {
		if failed != nil {
			logger.Info("pipeline deletion failed")
		} else {
			logger.Info("pipeline deletion progress unknown, failing pipeline")
		}

		statusCommand = FromState(pipeline)
	}

	return []PipelineCommand{
		*statusCommand,
		DeletePipelineWorkflows{
			Workflows: deletionWorkflows,
		},
	}
}

func (st PipelineStateHandler) onCreating(ctx context.Context, pipeline *pipelinesv1.Pipeline, creationWorkflows []argo.Workflow) []PipelineCommand {
	logger := log.FromContext(ctx)

	if pipeline.Status.Version == "" {
		failureMessage := "creating pipeline with empty version, failing pipeline"
		logger.Info(failureMessage)

		return []PipelineCommand{
			*FromState(pipeline).To(pipelinesv1.Failed).WithMessage(failureMessage),
		}
	}

	inProgress, succeeded, failed := latestWorkflowByPhase(creationWorkflows)

	if inProgress != nil {
		logger.V(2).Info("pipeline creation in progress")
		return []PipelineCommand{}
	}

	statusAfterCreating := func() *SetPipelineStatus {
		if succeeded == nil {
			var message string
			if failed != nil {
				message = "pipeline creation failed, failing pipeline"
			} else {
				message = "pipeline creation progress unknown, failing pipeline"
			}

			logger.Info(message)
			return FromState(pipeline).To(pipelinesv1.Failed).WithMessage(message)
		}

		idResult, err := getWorkflowOutput(succeeded, PipelineWorkflowConstants.PipelineIdParameterName)

		if err != nil {
			failureMessage := "could not retrieve kfpId, failing pipeline"
			logger.Error(err, failureMessage)
			return FromState(pipeline).To(pipelinesv1.Failed)
		}

		versionResult, _ := getWorkflowOutput(succeeded, PipelineWorkflowConstants.PipelineVersionParameterName)

		if versionResult == "" {
			message := "pipeline creation succeeded but version upload failed"
			logger.Info(message)
			return FromState(pipeline).
				To(pipelinesv1.Failed).
				WithKfpId(idResult).
				WithMessage(message)
		}

		logger.Info("pipeline creation succeeded")
		return FromState(pipeline).To(pipelinesv1.Succeeded).WithKfpId(idResult)
	}

	return []PipelineCommand{
		*statusAfterCreating(),
		DeletePipelineWorkflows{
			Workflows: creationWorkflows,
		},
	}
}
