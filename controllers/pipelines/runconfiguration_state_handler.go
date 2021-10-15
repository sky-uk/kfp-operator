package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"sort"
)

type RunConfigurationStateHandler struct {
	WorkflowFactory    RunConfigurationWorkflowFactory
	WorkflowRepository WorkflowRepository
}

func (st *RunConfigurationStateHandler) StateTransition(ctx context.Context, runconfiguration *pipelinesv1.RunConfiguration) []RunConfigurationCommand {

	if !runconfiguration.ObjectMeta.DeletionTimestamp.IsZero() &&
		(runconfiguration.Status.SynchronizationState == pipelinesv1.Succeeded ||
			runconfiguration.Status.SynchronizationState == pipelinesv1.Failed) {
		return st.onDelete(runconfiguration)
	}

	switch runconfiguration.Status.SynchronizationState {
	case pipelinesv1.Unknown:
		return st.onUnknown(runconfiguration)
	case pipelinesv1.Creating:
		return st.onCreating(runconfiguration,
			st.WorkflowRepository.GetByOperation(ctx,
				RunConfigurationWorkflowConstants.CreateOperationLabel,
				runconfiguration.NamespacedName(),
				RunConfigurationWorkflowConstants.RunConfigurationNameLabelKey))
	case pipelinesv1.Succeeded, pipelinesv1.Failed:
		return st.onSucceededOrFailed(runconfiguration)
	case pipelinesv1.Updating:
		return st.onUpdating(runconfiguration,
			st.WorkflowRepository.GetByOperation(ctx,
				RunConfigurationWorkflowConstants.UpdateOperationLabel,
				runconfiguration.NamespacedName(),
				RunConfigurationWorkflowConstants.RunConfigurationNameLabelKey))
	case pipelinesv1.Deleting:
		return st.onDeleting(runconfiguration,
			st.WorkflowRepository.GetByOperation(ctx,
				RunConfigurationWorkflowConstants.DeleteOperationLabel,
				runconfiguration.NamespacedName(),
				RunConfigurationWorkflowConstants.RunConfigurationNameLabelKey))
	case pipelinesv1.Deleted:
		return st.onDeleted()
	}

	return []RunConfigurationCommand{}
}

func (st *RunConfigurationStateHandler) onUnknown(runconfiguration *pipelinesv1.RunConfiguration) []RunConfigurationCommand {

	newRunconfigurationVersion := runconfiguration.Spec.ComputeVersion()

	if runconfiguration.Status.KfpId != "" {
		workflow := st.WorkflowFactory.ConstructUpdateWorkflow(runconfiguration)

		return []RunConfigurationCommand{
			CreateRunConfigurationWorkflow{Workflow: *workflow},
			SetRunConfigurationStatus{
				Status: pipelinesv1.Status{
					KfpId:                runconfiguration.Status.KfpId,
					Version:              newRunconfigurationVersion,
					SynchronizationState: pipelinesv1.Updating,
				},
			},
		}
	}

	workflow := st.WorkflowFactory.ConstructCreationWorkflow(runconfiguration)

	return []RunConfigurationCommand{
		CreateRunConfigurationWorkflow{Workflow: *workflow},
		SetRunConfigurationStatus{
			Status: pipelinesv1.Status{
				Version:              newRunconfigurationVersion,
				SynchronizationState: pipelinesv1.Creating,
			},
		},
	}
}

func (st RunConfigurationStateHandler) onDelete(runconfiguration *pipelinesv1.RunConfiguration) []RunConfigurationCommand {
	workflow := st.WorkflowFactory.ConstructDeletionWorkflow(runconfiguration)

	return []RunConfigurationCommand{
		CreateRunConfigurationWorkflow{Workflow: *workflow},
		SetRunConfigurationStatus{
			Status: pipelinesv1.Status{
				KfpId:                runconfiguration.Status.KfpId,
				Version:              runconfiguration.Status.Version,
				SynchronizationState: pipelinesv1.Deleting,
			},
		},
	}
}

func (st RunConfigurationStateHandler) onSucceededOrFailed(runconfiguration *pipelinesv1.RunConfiguration) []RunConfigurationCommand {
	newRunConfigurationVersion := runconfiguration.Spec.ComputeVersion()

	if runconfiguration.Status.Version == newRunConfigurationVersion {
		return []RunConfigurationCommand{}
	}

	var workflow *argo.Workflow
	var targetState pipelinesv1.SynchronizationState

	if runconfiguration.Status.KfpId == "" {
		workflow = st.WorkflowFactory.ConstructCreationWorkflow(runconfiguration)
		targetState = pipelinesv1.Creating
	} else {
		workflow = st.WorkflowFactory.ConstructUpdateWorkflow(runconfiguration)
		targetState = pipelinesv1.Updating
	}

	return []RunConfigurationCommand{
		CreateRunConfigurationWorkflow{Workflow: *workflow},
		SetRunConfigurationStatus{
			Status: pipelinesv1.Status{
				KfpId:                runconfiguration.Status.KfpId,
				Version:              newRunConfigurationVersion,
				SynchronizationState: targetState,
			},
		},
	}
}

func (st RunConfigurationStateHandler) onUpdating(runconfiguration *pipelinesv1.RunConfiguration, updateWorkflows []argo.Workflow) []RunConfigurationCommand {
	if runconfiguration.Status.Version == "" || runconfiguration.Status.KfpId == "" {
		return []RunConfigurationCommand{
			SetRunConfigurationStatus{
				Status: pipelinesv1.Status{
					Version:              runconfiguration.Status.Version,
					KfpId:                runconfiguration.Status.KfpId,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	inProgress, succeeded, _ := latestWorkflowByPhase(updateWorkflows)

	if inProgress != nil {
		return []RunConfigurationCommand{}
	}

	newStatus := runconfiguration.Status.DeepCopy()

	if succeeded != nil {
		newStatus.SynchronizationState = pipelinesv1.Succeeded
	} else {
		newStatus.SynchronizationState = pipelinesv1.Failed
	}

	return []RunConfigurationCommand{
		SetRunConfigurationStatus{
			Status: *newStatus,
		},
		DeleteRunConfigurationWorkflows{
			Workflows: updateWorkflows,
		},
	}
}

func (st RunConfigurationStateHandler) onDeleting(runconfiguration *pipelinesv1.RunConfiguration, deletionWorkflows []argo.Workflow) []RunConfigurationCommand {

	inProgress, succeeded, _ := latestWorkflowByPhase(deletionWorkflows)

	if inProgress != nil {
		return []RunConfigurationCommand{}
	}

	newStatus := runconfiguration.Status.DeepCopy()

	if succeeded != nil {
		newStatus.SynchronizationState = pipelinesv1.Deleted
	}

	return []RunConfigurationCommand{
		DeleteRunConfigurationWorkflows{
			Workflows: deletionWorkflows,
		},
		SetRunConfigurationStatus{
			Status: *newStatus,
		},
	}
}

func (st RunConfigurationStateHandler) onDeleted() []RunConfigurationCommand {
	return []RunConfigurationCommand{
		DeleteRunConfiguration{},
	}
}

func (st RunConfigurationStateHandler) onCreating(runconfiguration *pipelinesv1.RunConfiguration, creationWorkflows []argo.Workflow) []RunConfigurationCommand {
	if runconfiguration.Status.Version == "" {
		return []RunConfigurationCommand{
			SetRunConfigurationStatus{
				Status: pipelinesv1.Status{
					KfpId:                runconfiguration.Status.KfpId,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	sort.Slice(creationWorkflows, func(i, j int) bool {
		return creationWorkflows[i].ObjectMeta.CreationTimestamp.Before(&creationWorkflows[j].ObjectMeta.CreationTimestamp)
	})

	inProgress, succeeded, _ := latestWorkflowByPhase(creationWorkflows)

	if inProgress != nil {
		return []RunConfigurationCommand{}
	}

	newStatus := runconfiguration.Status.DeepCopy()

	if succeeded != nil {
		newStatus.SynchronizationState = pipelinesv1.Succeeded
		idResult, error := getWorkflowOutput(succeeded, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)

		if error != nil {
			newStatus.SynchronizationState = pipelinesv1.Failed
		} else {
			newStatus.KfpId = idResult
		}
	} else {
		newStatus.SynchronizationState = pipelinesv1.Failed
	}

	return []RunConfigurationCommand{
		SetRunConfigurationStatus{
			Status: *newStatus,
		},
		DeleteRunConfigurationWorkflows{
			Workflows: creationWorkflows,
		},
	}
}
