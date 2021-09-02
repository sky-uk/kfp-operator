package controllers

import (
	"sort"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/api/v1"
)

func stateTransition(pipeline *pipelinesv1.Pipeline, workflows Workflows) []Command {
	if !pipeline.ObjectMeta.DeletionTimestamp.IsZero() &&
		(pipeline.Status.SynchronizationState == pipelinesv1.Succeeded ||
			pipeline.Status.SynchronizationState == pipelinesv1.Failed) {
		return onDelete(pipeline)
	}

	switch pipeline.Status.SynchronizationState {
	case pipelinesv1.Unknown:
		return onUnknown(pipeline)
	case pipelinesv1.Creating:
		return onCreating(pipeline, workflows.GetByOperation(Create))
	case pipelinesv1.Succeeded:
		return onSucceeded(pipeline)
	case pipelinesv1.Updating:
		return onUpdating(pipeline, workflows.GetByOperation(Update))
	case pipelinesv1.Deleting:
		return onDeleting(pipeline, workflows.GetByOperation(Delete))
	case pipelinesv1.Deleted:
		return onDeleted(pipeline)
	}

	return []Command{}
}

func onUnknown(pipeline *pipelinesv1.Pipeline) []Command {

	if pipeline.Status.Id != "" {
		workflow, error := constructUpdateWorkflow(pipeline)

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

		newPipelineVersion := pipelinesv1.ComputeVersion(pipeline.Spec)

		return []Command{
			SetPipelineStatus{
				Status: pipelinesv1.PipelineStatus{
					Id:                   pipeline.Status.Id,
					Version:              newPipelineVersion,
					SynchronizationState: pipelinesv1.Updating,
				},
			},
			CreateWorkflow{Workflow: *workflow},
		}
	}

	pipelineVersion := pipelinesv1.ComputeVersion(pipeline.Spec)

	workflow, error := constructCreationWorkflow(pipeline)

	if error != nil {
		return []Command{
			SetPipelineStatus{
				Status: pipelinesv1.PipelineStatus{
					Version:              pipelineVersion,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	return []Command{
		SetPipelineStatus{
			Status: pipelinesv1.PipelineStatus{
				Version:              pipelineVersion,
				SynchronizationState: pipelinesv1.Creating,
			},
		},
		CreateWorkflow{Workflow: *workflow},
	}
}

func onDelete(pipeline *pipelinesv1.Pipeline) []Command {
	workflow := constructDeletionWorkflow(pipeline)

	return []Command{
		SetPipelineStatus{
			Status: pipelinesv1.PipelineStatus{
				Id:                   pipeline.Status.Id,
				Version:              pipeline.Status.Version,
				SynchronizationState: pipelinesv1.Deleting,
			},
		},
		CreateWorkflow{Workflow: *workflow},
	}
}

func onSucceeded(pipeline *pipelinesv1.Pipeline) []Command {
	newPipelineVersion := pipelinesv1.ComputeVersion(pipeline.Spec)

	if pipeline.Status.Version == newPipelineVersion {
		return []Command{}
	}

	// TODO check
	workflow, error := constructUpdateWorkflow(pipeline)

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
		SetPipelineStatus{
			Status: pipelinesv1.PipelineStatus{
				Id:                   pipeline.Status.Id,
				Version:              newPipelineVersion,
				SynchronizationState: pipelinesv1.Updating,
			},
		},
		CreateWorkflow{Workflow: *workflow},
	}
}

func onUpdating(pipeline *pipelinesv1.Pipeline, updateWorkflows []argo.Workflow) []Command {
	if pipeline.Status.Version == "" || pipeline.Status.Id == "" {

		return []Command{
			SetPipelineStatus{
				Status: pipelinesv1.PipelineStatus{
					Version:              pipeline.Status.Version,
					Id:                   pipeline.Status.Id,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	inProgress, succeeded, _ := latestWorkflowsByPhase(updateWorkflows)

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

func onDeleting(pipeline *pipelinesv1.Pipeline, deletionWorkflows []argo.Workflow) []Command {

	inProgress, succeeded, _ := latestWorkflowsByPhase(deletionWorkflows)

	if inProgress != nil {
		return []Command{}
	}

	newStatus := pipeline.Status.DeepCopy()

	if succeeded != nil {
		newStatus.SynchronizationState = pipelinesv1.Deleted
	}

	return []Command{
		SetPipelineStatus{
			Status: *newStatus,
		},
		DeleteWorkflows{
			Workflows: deletionWorkflows,
		},
	}
}

func onDeleted(pipeline *pipelinesv1.Pipeline) []Command {

	return []Command{
		DeletePipeline{},
	}
}

func onCreating(pipeline *pipelinesv1.Pipeline, creationWorkflows []argo.Workflow) []Command {
	if pipeline.Status.Version == "" {
		return []Command{
			SetPipelineStatus{
				Status: pipelinesv1.PipelineStatus{
					Id:                   pipeline.Status.Id,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	sort.Slice(creationWorkflows, func(i, j int) bool {
		return creationWorkflows[i].ObjectMeta.CreationTimestamp.Before(&creationWorkflows[j].ObjectMeta.CreationTimestamp)
	})

	inProgress, succeeded, failed := latestWorkflowsByPhase(creationWorkflows)

	if inProgress != nil {
		return []Command{}
	}

	newStatus := pipeline.Status.DeepCopy()

	if succeeded != nil {
		newStatus.SynchronizationState = pipelinesv1.Succeeded
		newStatus.Id = workflowOutput(*succeeded, PipelineIdKey)
	} else {
		if failed != nil {
			newStatus.Id = workflowOutput(*failed, PipelineIdKey)
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
