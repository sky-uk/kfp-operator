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
			logger.Error(err, "error constructing update workflow, failing pipeline")

			return []PipelineCommand{
				SetPipelineStatus{
					Status: pipelinesv1.Status{
						Version:              pipeline.Status.Version,
						SynchronizationState: pipelinesv1.Failed,
					},
				},
			}
		}

		return []PipelineCommand{
			SetPipelineStatus{
				Status: pipelinesv1.Status{
					KfpId:                pipeline.Status.KfpId,
					Version:              newPipelineVersion,
					SynchronizationState: pipelinesv1.Updating,
				},
			},
			CreatePipelineWorkflow{Workflow: *workflow},
		}
	}

	logger.Info("empty state, creating pipeline")
	workflow, err := st.WorkflowFactory.ConstructCreationWorkflow(ctx, pipeline)

	if err != nil {
		logger.Error(err, "error constructing creation workflow, failing pipeline")

		return []PipelineCommand{
			SetPipelineStatus{
				Status: pipelinesv1.Status{
					Version:              newPipelineVersion,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	return []PipelineCommand{
		SetPipelineStatus{
			Status: pipelinesv1.Status{
				Version:              newPipelineVersion,
				SynchronizationState: pipelinesv1.Creating,
			},
		},
		CreatePipelineWorkflow{Workflow: *workflow},
	}
}

func (st PipelineStateHandler) onDelete(ctx context.Context, pipeline *pipelinesv1.Pipeline) []PipelineCommand {
	logger := log.FromContext(ctx)
	logger.Info("deletion requested, deleting")
	workflow := st.WorkflowFactory.ConstructDeletionWorkflow(ctx, pipeline)

	return []PipelineCommand{
		SetPipelineStatus{
			Status: pipelinesv1.Status{
				KfpId:                pipeline.Status.KfpId,
				Version:              pipeline.Status.Version,
				SynchronizationState: pipelinesv1.Deleting,
			},
		},
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
			SetPipelineStatus{
				Status: pipelinesv1.Status{
					KfpId:                pipeline.Status.KfpId,
					Version:              newPipelineVersion,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	return []PipelineCommand{
		SetPipelineStatus{
			Status: pipelinesv1.Status{
				KfpId:                pipeline.Status.KfpId,
				Version:              newPipelineVersion,
				SynchronizationState: targetState,
			},
		},
		CreatePipelineWorkflow{Workflow: *workflow},
	}
}

func (st PipelineStateHandler) onUpdating(ctx context.Context, pipeline *pipelinesv1.Pipeline, updateWorkflows []argo.Workflow) []PipelineCommand {
	logger := log.FromContext(ctx)

	if pipeline.Status.Version == "" || pipeline.Status.KfpId == "" {
		logger.Info("updating pipeline with empty version or kfpId, failing pipeline")
		return []PipelineCommand{
			SetPipelineStatus{
				Status: pipelinesv1.Status{
					Version:              pipeline.Status.Version,
					KfpId:                pipeline.Status.KfpId,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	inProgress, succeeded, failed := latestWorkflowByPhase(updateWorkflows)

	if inProgress != nil {
		logger.V(2).Info("pipeline update in progress")
		return []PipelineCommand{}
	}

	newStatus := pipeline.Status.DeepCopy()

	if succeeded != nil {
		logger.Info("pipeline update succeeded")
		newStatus.SynchronizationState = pipelinesv1.Succeeded
	} else {
		if failed != nil {
			logger.Info("pipeline update failed")
		} else {
			logger.Info("pipeline updating progress unknown, failing pipeline")
		}
		newStatus.SynchronizationState = pipelinesv1.Failed
	}

	return []PipelineCommand{
		SetPipelineStatus{
			Status: *newStatus,
		},
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

	newStatus := pipeline.Status.DeepCopy()

	if succeeded != nil {
		logger.Info("pipeline deletion succeeded")
		newStatus.SynchronizationState = pipelinesv1.Deleted
	} else if failed != nil {
		logger.Info("pipeline deletion failed")
	} else {
		logger.Info("pipeline deletion progress unknown, failing pipeline")
	}

	return []PipelineCommand{
		SetPipelineStatus{
			Status: *newStatus,
		},
		DeletePipelineWorkflows{
			Workflows: deletionWorkflows,
		},
	}
}

func (st PipelineStateHandler) onCreating(ctx context.Context, pipeline *pipelinesv1.Pipeline, creationWorkflows []argo.Workflow) []PipelineCommand {
	logger := log.FromContext(ctx)

	if pipeline.Status.Version == "" {
		logger.Info("creating pipeline with empty version, failing pipeline")
		return []PipelineCommand{
			SetPipelineStatus{
				Status: pipelinesv1.Status{
					KfpId:                pipeline.Status.KfpId,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	inProgress, succeeded, failed := latestWorkflowByPhase(creationWorkflows)

	if inProgress != nil {
		logger.V(2).Info("pipeline creation in progress")
		return []PipelineCommand{}
	}

	statusAfterCreating := func() (newStatus pipelinesv1.Status) {
		newStatus = *pipeline.Status.DeepCopy()

		if succeeded == nil {
			if failed != nil {
				logger.Info("pipeline creation failed, failing pipeline")
			} else {
				logger.Info("pipeline creation progress unknown, failing pipeline")
			}

			newStatus.SynchronizationState = pipelinesv1.Failed
			return
		}

		idResult, err := getWorkflowOutput(succeeded, PipelineWorkflowConstants.PipelineIdParameterName)

		if err != nil {
			logger.Error(err, "could not retrieve kfpId, failing pipeline")
			newStatus.SynchronizationState = pipelinesv1.Failed
			return
		}

		newStatus.KfpId = idResult
		versionResult, _ := getWorkflowOutput(succeeded, PipelineWorkflowConstants.PipelineVersionParameterName)

		if versionResult == "" {
			logger.Info("pipeline creation succeeded but version upload failed")
			newStatus.SynchronizationState = pipelinesv1.Failed
			return
		}

		logger.Info("pipeline creation succeeded")
		newStatus.SynchronizationState = pipelinesv1.Succeeded
		return
	}

	return []PipelineCommand{
		SetPipelineStatus{
			Status: statusAfterCreating(),
		},
		DeletePipelineWorkflows{
			Workflows: creationWorkflows,
		},
	}
}
