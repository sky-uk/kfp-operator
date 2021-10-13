package pipelines

import (
	"context"
	"sort"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

type StateHandler struct {
	WorkflowFactory    PipelineWorkflowFactory
	WorkflowRepository WorkflowRepository
}

func (st StateHandler) StateTransition(ctx context.Context, pipeline *pipelinesv1.Pipeline) []Command {

	if !pipeline.ObjectMeta.DeletionTimestamp.IsZero() &&
		(pipeline.Status.SynchronizationState == pipelinesv1.Succeeded ||
			pipeline.Status.SynchronizationState == pipelinesv1.Failed) {
		return st.onDelete(pipeline)
	}

	switch pipeline.Status.SynchronizationState {
	case pipelinesv1.Unknown:
		return st.onUnknown(pipeline)
	case pipelinesv1.Creating:
		return st.onCreating(pipeline, st.WorkflowRepository.GetByOperation(ctx, PipelineWorkflowConstants.CreateOperationLabel, pipeline))
	case pipelinesv1.Succeeded, pipelinesv1.Failed:
		return st.onSucceededOrFailed(pipeline)
	case pipelinesv1.Updating:
		return st.onUpdating(pipeline, st.WorkflowRepository.GetByOperation(ctx, PipelineWorkflowConstants.UpdateOperationLabel, pipeline))
	case pipelinesv1.Deleting:
		return st.onDeleting(pipeline, st.WorkflowRepository.GetByOperation(ctx, PipelineWorkflowConstants.DeleteOperationLabel, pipeline))
	case pipelinesv1.Deleted:
		return st.onDeleted(pipeline)
	}

	return []Command{}
}

func (st StateHandler) onUnknown(pipeline *pipelinesv1.Pipeline) []Command {

	newPipelineVersion := pipeline.Spec.ComputeVersion()

	if pipeline.Status.KfpId != "" {
		workflow, error := st.WorkflowFactory.ConstructUpdateWorkflow(pipeline.Spec, pipeline.ObjectMeta, pipeline.Status.KfpId, newPipelineVersion)

		if error != nil {
			return []Command{
				SetPipelineStatus{
					Status: pipelinesv1.PipelineStatus{
						Version:              pipeline.Status.Version,
						SynchronizationState: pipelinesv1.Failed,
					},
				},
			}
		}

		return []Command{
			CreateWorkflow{Workflow: *workflow},
			SetPipelineStatus{
				Status: pipelinesv1.PipelineStatus{
					KfpId:                pipeline.Status.KfpId,
					Version:              newPipelineVersion,
					SynchronizationState: pipelinesv1.Updating,
				},
			},
		}
	}

	workflow, error := st.WorkflowFactory.ConstructCreationWorkflow(pipeline.Spec, pipeline.ObjectMeta, newPipelineVersion)

	if error != nil {
		return []Command{
			SetPipelineStatus{
				Status: pipelinesv1.PipelineStatus{
					Version:              newPipelineVersion,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	return []Command{
		CreateWorkflow{Workflow: *workflow},
		SetPipelineStatus{
			Status: pipelinesv1.PipelineStatus{
				Version:              newPipelineVersion,
				SynchronizationState: pipelinesv1.Creating,
			},
		},
	}
}

func (st StateHandler) onDelete(pipeline *pipelinesv1.Pipeline) []Command {
	workflow := st.WorkflowFactory.ConstructDeletionWorkflow(pipeline.ObjectMeta, pipeline.Status.KfpId)

	return []Command{
		CreateWorkflow{Workflow: *workflow},
		SetPipelineStatus{
			Status: pipelinesv1.PipelineStatus{
				KfpId:                pipeline.Status.KfpId,
				Version:              pipeline.Status.Version,
				SynchronizationState: pipelinesv1.Deleting,
			},
		},
	}
}

func (st StateHandler) onSucceededOrFailed(pipeline *pipelinesv1.Pipeline) []Command {
	newPipelineVersion := pipeline.Spec.ComputeVersion()

	if pipeline.Status.Version == newPipelineVersion {
		return []Command{}
	}

	var workflow *argo.Workflow
	var error error
	var targetState pipelinesv1.SynchronizationState

	if pipeline.Status.KfpId == "" {
		workflow, error = st.WorkflowFactory.ConstructCreationWorkflow(pipeline.Spec, pipeline.ObjectMeta, newPipelineVersion)
		targetState = pipelinesv1.Creating
	} else {
		workflow, error = st.WorkflowFactory.ConstructUpdateWorkflow(pipeline.Spec, pipeline.ObjectMeta, pipeline.Status.KfpId, newPipelineVersion)
		targetState = pipelinesv1.Updating
	}

	if error != nil {
		return []Command{
			SetPipelineStatus{
				Status: pipelinesv1.PipelineStatus{
					KfpId:                pipeline.Status.KfpId,
					Version:              newPipelineVersion,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	return []Command{
		CreateWorkflow{Workflow: *workflow},
		SetPipelineStatus{
			Status: pipelinesv1.PipelineStatus{
				KfpId:                pipeline.Status.KfpId,
				Version:              newPipelineVersion,
				SynchronizationState: targetState,
			},
		},
	}
}

func (st StateHandler) onUpdating(pipeline *pipelinesv1.Pipeline, updateWorkflows []argo.Workflow) []Command {
	if pipeline.Status.Version == "" || pipeline.Status.KfpId == "" {
		return []Command{
			SetPipelineStatus{
				Status: pipelinesv1.PipelineStatus{
					Version:              pipeline.Status.Version,
					KfpId:                pipeline.Status.KfpId,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	inProgress, succeeded, _ := latestWorkflowByPhase(updateWorkflows)

	if inProgress != nil {
		return []Command{}
	}

	newStatus := pipeline.Status.DeepCopy()

	if succeeded != nil {
		newStatus.SynchronizationState = pipelinesv1.Succeeded
	} else {
		newStatus.SynchronizationState = pipelinesv1.Failed
	}

	return []Command{
		SetPipelineStatus{
			Status: *newStatus,
		},
		DeleteWorkflows{
			Workflows: updateWorkflows,
		},
	}
}

func (st StateHandler) onDeleting(pipeline *pipelinesv1.Pipeline, deletionWorkflows []argo.Workflow) []Command {

	inProgress, succeeded, _ := latestWorkflowByPhase(deletionWorkflows)

	if inProgress != nil {
		return []Command{}
	}

	newStatus := pipeline.Status.DeepCopy()

	if succeeded != nil {
		newStatus.SynchronizationState = pipelinesv1.Deleted
	}

	return []Command{
		DeleteWorkflows{
			Workflows: deletionWorkflows,
		},
		SetPipelineStatus{
			Status: *newStatus,
		},
	}
}

func (st StateHandler) onDeleted(pipeline *pipelinesv1.Pipeline) []Command {
	return []Command{
		DeletePipeline{},
	}
}

func (st StateHandler) onCreating(pipeline *pipelinesv1.Pipeline, creationWorkflows []argo.Workflow) []Command {

	if pipeline.Status.Version == "" {
		return []Command{
			SetPipelineStatus{
				Status: pipelinesv1.PipelineStatus{
					KfpId:                pipeline.Status.KfpId,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	sort.Slice(creationWorkflows, func(i, j int) bool {
		return creationWorkflows[i].ObjectMeta.CreationTimestamp.Before(&creationWorkflows[j].ObjectMeta.CreationTimestamp)
	})

	inProgress, succeeded, failed := latestWorkflowByPhase(creationWorkflows)

	if inProgress != nil {
		return []Command{}
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

	return []Command{
		SetPipelineStatus{
			Status: *newStatus,
		},
		DeleteWorkflows{
			Workflows: creationWorkflows,
		},
	}
}
