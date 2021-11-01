package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

type PipelineStateHandler struct {
	WorkflowFactory    PipelineWorkflowFactory
	WorkflowRepository WorkflowRepository
}

func (st PipelineStateHandler) StateTransition(ctx context.Context, pipeline *pipelinesv1.Pipeline) []PipelineCommand {

	if !pipeline.ObjectMeta.DeletionTimestamp.IsZero() &&
		(pipeline.Status.SynchronizationState == pipelinesv1.Succeeded ||
			pipeline.Status.SynchronizationState == pipelinesv1.Failed) {
		return st.onDelete(ctx, pipeline)
	}

	switch pipeline.Status.SynchronizationState {
	case pipelinesv1.Unknown:
		return st.onUnknown(ctx, pipeline)
	case pipelinesv1.Creating:
		return st.onCreating(pipeline,
			st.WorkflowRepository.GetByOperation(ctx,
				PipelineWorkflowConstants.CreateOperationLabel,
				pipeline.NamespacedName(),
				PipelineWorkflowConstants.PipelineNameLabelKey))
	case pipelinesv1.Succeeded, pipelinesv1.Failed:
		return st.onSucceededOrFailed(ctx, pipeline)
	case pipelinesv1.Updating:
		return st.onUpdating(pipeline,
			st.WorkflowRepository.GetByOperation(ctx,
				PipelineWorkflowConstants.UpdateOperationLabel,
				pipeline.NamespacedName(),
				PipelineWorkflowConstants.PipelineNameLabelKey))
	case pipelinesv1.Deleting:
		return st.onDeleting(pipeline,
			st.WorkflowRepository.GetByOperation(ctx,
				PipelineWorkflowConstants.DeleteOperationLabel,
				pipeline.NamespacedName(),
				PipelineWorkflowConstants.PipelineNameLabelKey))
	case pipelinesv1.Deleted:
		return st.onDeleted()
	}

	return []PipelineCommand{}
}

func (st PipelineStateHandler) onUnknown(ctx context.Context, pipeline *pipelinesv1.Pipeline) []PipelineCommand {

	newPipelineVersion := pipeline.Spec.ComputeVersion()

	if pipeline.Status.KfpId != "" {
		workflow, error := st.WorkflowFactory.ConstructUpdateWorkflow(ctx, pipeline)

		if error != nil {
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
			CreatePipelineWorkflow{Workflow: *workflow},
			SetPipelineStatus{
				Status: pipelinesv1.Status{
					KfpId:                pipeline.Status.KfpId,
					Version:              newPipelineVersion,
					SynchronizationState: pipelinesv1.Updating,
				},
			},
		}
	}

	workflow, error := st.WorkflowFactory.ConstructCreationWorkflow(ctx, pipeline)

	if error != nil {
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
		CreatePipelineWorkflow{Workflow: *workflow},
		SetPipelineStatus{
			Status: pipelinesv1.Status{
				Version:              newPipelineVersion,
				SynchronizationState: pipelinesv1.Creating,
			},
		},
	}
}

func (st PipelineStateHandler) onDelete(ctx context.Context, pipeline *pipelinesv1.Pipeline) []PipelineCommand {
	workflow := st.WorkflowFactory.ConstructDeletionWorkflow(ctx, pipeline)

	return []PipelineCommand{
		CreatePipelineWorkflow{Workflow: *workflow},
		SetPipelineStatus{
			Status: pipelinesv1.Status{
				KfpId:                pipeline.Status.KfpId,
				Version:              pipeline.Status.Version,
				SynchronizationState: pipelinesv1.Deleting,
			},
		},
	}
}

func (st PipelineStateHandler) onSucceededOrFailed(ctx context.Context, pipeline *pipelinesv1.Pipeline) []PipelineCommand {
	newPipelineVersion := pipeline.Spec.ComputeVersion()

	if pipeline.Status.Version == newPipelineVersion {
		return []PipelineCommand{}
	}

	var workflow *argo.Workflow
	var error error
	var targetState pipelinesv1.SynchronizationState

	if pipeline.Status.KfpId == "" {
		workflow, error = st.WorkflowFactory.ConstructCreationWorkflow(ctx, pipeline)
		targetState = pipelinesv1.Creating
	} else {
		workflow, error = st.WorkflowFactory.ConstructUpdateWorkflow(ctx, pipeline)
		targetState = pipelinesv1.Updating
	}

	if error != nil {
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
		CreatePipelineWorkflow{Workflow: *workflow},
		SetPipelineStatus{
			Status: pipelinesv1.Status{
				KfpId:                pipeline.Status.KfpId,
				Version:              newPipelineVersion,
				SynchronizationState: targetState,
			},
		},
	}
}

func (st PipelineStateHandler) onUpdating(pipeline *pipelinesv1.Pipeline, updateWorkflows []argo.Workflow) []PipelineCommand {
	if pipeline.Status.Version == "" || pipeline.Status.KfpId == "" {
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

	inProgress, succeeded, _ := latestWorkflowByPhase(updateWorkflows)

	if inProgress != nil {
		return []PipelineCommand{}
	}

	newStatus := pipeline.Status.DeepCopy()

	if succeeded != nil {
		newStatus.SynchronizationState = pipelinesv1.Succeeded
	} else {
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

func (st PipelineStateHandler) onDeleting(pipeline *pipelinesv1.Pipeline, deletionWorkflows []argo.Workflow) []PipelineCommand {

	inProgress, succeeded, _ := latestWorkflowByPhase(deletionWorkflows)

	if inProgress != nil {
		return []PipelineCommand{}
	}

	newStatus := pipeline.Status.DeepCopy()

	if succeeded != nil {
		newStatus.SynchronizationState = pipelinesv1.Deleted
	}

	return []PipelineCommand{
		DeletePipelineWorkflows{
			Workflows: deletionWorkflows,
		},
		SetPipelineStatus{
			Status: *newStatus,
		},
	}
}

func (st PipelineStateHandler) onDeleted() []PipelineCommand {
	return []PipelineCommand{
		DeletePipeline{},
	}
}

func (st PipelineStateHandler) onCreating(pipeline *pipelinesv1.Pipeline, creationWorkflows []argo.Workflow) []PipelineCommand {
	if pipeline.Status.Version == "" {
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
		return []PipelineCommand{}
	}

	newStatus := pipeline.Status.DeepCopy()

	if succeeded != nil {
		newStatus.SynchronizationState = pipelinesv1.Succeeded
		idResult, error := getWorkflowOutput(succeeded, PipelineWorkflowConstants.PipelineIdParameterName)

		if error != nil {
			newStatus.SynchronizationState = pipelinesv1.Failed
		} else {
			newStatus.KfpId = idResult
		}
	} else {
		if failed != nil {
			idResult, error := getWorkflowOutput(failed, PipelineWorkflowConstants.PipelineIdParameterName)

			if error == nil {
				newStatus.KfpId = idResult
			}
		}

		newStatus.SynchronizationState = pipelinesv1.Failed
	}

	return []PipelineCommand{
		SetPipelineStatus{
			Status: *newStatus,
		},
		DeletePipelineWorkflows{
			Workflows: creationWorkflows,
		},
	}
}
