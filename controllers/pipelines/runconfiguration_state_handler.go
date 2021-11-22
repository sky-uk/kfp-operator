package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type RunConfigurationStateHandler struct {
	WorkflowFactory    RunConfigurationWorkflowFactory
	WorkflowRepository WorkflowRepository
}

func (st *RunConfigurationStateHandler) StateTransition(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) []RunConfigurationCommand {
	logger := log.FromContext(ctx)
	logger.Info("state transition start")

	if !runConfiguration.ObjectMeta.DeletionTimestamp.IsZero() &&
		(runConfiguration.Status.SynchronizationState == pipelinesv1.Succeeded ||
			runConfiguration.Status.SynchronizationState == pipelinesv1.Failed) {
		return st.onDelete(ctx, runConfiguration)
	}

	switch runConfiguration.Status.SynchronizationState {
	case pipelinesv1.Creating:
		return st.onCreating(ctx, runConfiguration,
			st.WorkflowRepository.GetByOperation(ctx,
				RunConfigurationWorkflowConstants.CreateOperationLabel,
				runConfiguration.NamespacedName(),
				RunConfigurationWorkflowConstants.RunConfigurationNameLabelKey))
	case pipelinesv1.Succeeded, pipelinesv1.Failed:
		return st.onSucceededOrFailed(ctx, runConfiguration)
	case pipelinesv1.Updating:
		return st.onUpdating(ctx, runConfiguration,
			st.WorkflowRepository.GetByOperation(ctx,
				RunConfigurationWorkflowConstants.UpdateOperationLabel,
				runConfiguration.NamespacedName(),
				RunConfigurationWorkflowConstants.RunConfigurationNameLabelKey))
	case pipelinesv1.Deleting:
		return st.onDeleting(ctx, runConfiguration,
			st.WorkflowRepository.GetByOperation(ctx,
				RunConfigurationWorkflowConstants.DeleteOperationLabel,
				runConfiguration.NamespacedName(),
				RunConfigurationWorkflowConstants.RunConfigurationNameLabelKey))
	case pipelinesv1.Deleted:
		return st.onDeleted()
	default:
		return st.onUnknown(ctx, runConfiguration)
	}
}

func (st *RunConfigurationStateHandler) onUnknown(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) []RunConfigurationCommand {
	logger := log.FromContext(ctx)

	newRunconfigurationVersion := runConfiguration.Spec.ComputeVersion()

	if runConfiguration.Status.KfpId != "" {
		logger.Info("empty state but KfpId already exists, updating run configuration")
		workflow := st.WorkflowFactory.ConstructUpdateWorkflow(ctx, runConfiguration)

		return []RunConfigurationCommand{
			SetRunConfigurationStatus{
				Status: pipelinesv1.Status{
					KfpId:                runConfiguration.Status.KfpId,
					Version:              newRunconfigurationVersion,
					SynchronizationState: pipelinesv1.Updating,
				},
			},
			CreateRunConfigurationWorkflow{Workflow: *workflow},
		}
	}

	logger.Info("empty state, creating run configuration")
	workflow := st.WorkflowFactory.ConstructCreationWorkflow(ctx, runConfiguration)

	return []RunConfigurationCommand{
		SetRunConfigurationStatus{
			Status: pipelinesv1.Status{
				Version:              newRunconfigurationVersion,
				SynchronizationState: pipelinesv1.Creating,
			},
		},
		CreateRunConfigurationWorkflow{Workflow: *workflow},
	}
}

func (st RunConfigurationStateHandler) onDelete(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) []RunConfigurationCommand {
	logger := log.FromContext(ctx)
	logger.Info("deletion requested, deleting")
	workflow := st.WorkflowFactory.ConstructDeletionWorkflow(ctx, runConfiguration)

	return []RunConfigurationCommand{
		SetRunConfigurationStatus{
			Status: pipelinesv1.Status{
				KfpId:                runConfiguration.Status.KfpId,
				Version:              runConfiguration.Status.Version,
				SynchronizationState: pipelinesv1.Deleting,
			},
		},
		CreateRunConfigurationWorkflow{Workflow: *workflow},
	}
}

func (st RunConfigurationStateHandler) onSucceededOrFailed(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) []RunConfigurationCommand {
	logger := log.FromContext(ctx)
	newRunConfigurationVersion := runConfiguration.Spec.ComputeVersion()

	if runConfiguration.Status.Version == newRunConfigurationVersion {
		logger.V(2).Info("run configuration version has not changed")
		return []RunConfigurationCommand{}
	}

	var workflow *argo.Workflow
	var targetState pipelinesv1.SynchronizationState

	if runConfiguration.Status.KfpId == "" {
		logger.Info("no kfpId exists, creating")
		workflow = st.WorkflowFactory.ConstructCreationWorkflow(ctx, runConfiguration)
		targetState = pipelinesv1.Creating
	} else {
		logger.Info("kfpId exists, updating")
		workflow = st.WorkflowFactory.ConstructUpdateWorkflow(ctx, runConfiguration)
		targetState = pipelinesv1.Updating
	}

	return []RunConfigurationCommand{
		SetRunConfigurationStatus{
			Status: pipelinesv1.Status{
				KfpId:                runConfiguration.Status.KfpId,
				Version:              newRunConfigurationVersion,
				SynchronizationState: targetState,
			},
		},
		CreateRunConfigurationWorkflow{Workflow: *workflow},
	}
}

func (st RunConfigurationStateHandler) onUpdating(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration, updateWorkflows []argo.Workflow) []RunConfigurationCommand {
	logger := log.FromContext(ctx)

	if runConfiguration.Status.Version == "" || runConfiguration.Status.KfpId == "" {
		logger.Info("updating run configuration with empty version or kfpId, failing run configuration")
		return []RunConfigurationCommand{
			SetRunConfigurationStatus{
				Status: pipelinesv1.Status{
					Version:              runConfiguration.Status.Version,
					KfpId:                runConfiguration.Status.KfpId,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	inProgress, succeeded, failed := latestWorkflowByPhase(updateWorkflows)

	if inProgress != nil {
		logger.V(2).Info("run configuration update in progress")
		return []RunConfigurationCommand{}
	}

	newStatus := runConfiguration.Status.DeepCopy()

	if succeeded != nil {
		logger.Info("run configuration update succeeded")
		newStatus.SynchronizationState = pipelinesv1.Succeeded
		idResult, err := getWorkflowOutput(succeeded, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)

		if err != nil {
			logger.Error(err, "could not retrieve workflow output, failing run configuration")
			newStatus.SynchronizationState = pipelinesv1.Failed
		} else {
			newStatus.KfpId = idResult
		}
	} else {
		if failed != nil {
			logger.Info("run configuration update failed")
		} else {
			logger.Info("run configuration creation progress unknown, failing run configuration")
		}
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

func (st RunConfigurationStateHandler) onDeleting(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration, deletionWorkflows []argo.Workflow) []RunConfigurationCommand {
	logger := log.FromContext(ctx)

	inProgress, succeeded, failed := latestWorkflowByPhase(deletionWorkflows)

	if inProgress != nil {
		logger.V(2).Info("run configuration deletion in progress")
		return []RunConfigurationCommand{}
	}

	newStatus := runConfiguration.Status.DeepCopy()

	if succeeded != nil {
		logger.Info("run configuration deletion succeeded")
		newStatus.SynchronizationState = pipelinesv1.Deleted
	} else if failed != nil {
		logger.Info("run configuration deletion failed")
	} else {
		logger.Info("run configuration deletion progress unknown, failing run configuration")
	}

	return []RunConfigurationCommand{
		SetRunConfigurationStatus{
			Status: *newStatus,
		},
		DeleteRunConfigurationWorkflows{
			Workflows: deletionWorkflows,
		},
	}
}

func (st RunConfigurationStateHandler) onDeleted() []RunConfigurationCommand {
	return []RunConfigurationCommand{
		DeleteRunConfiguration{},
	}
}

func (st RunConfigurationStateHandler) onCreating(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration, creationWorkflows []argo.Workflow) []RunConfigurationCommand {
	logger := log.FromContext(ctx)

	if runConfiguration.Status.Version == "" {
		logger.Info("creating run configuration with empty version, failing run configuration")
		return []RunConfigurationCommand{
			SetRunConfigurationStatus{
				Status: pipelinesv1.Status{
					KfpId:                runConfiguration.Status.KfpId,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	inProgress, succeeded, failed := latestWorkflowByPhase(creationWorkflows)

	if inProgress != nil {
		logger.V(2).Info("run configuration creation in progress")
		return []RunConfigurationCommand{}
	}

	newStatus := runConfiguration.Status.DeepCopy()

	if succeeded != nil {
		logger.Info("run configuration creation succeeded")
		newStatus.SynchronizationState = pipelinesv1.Succeeded
		idResult, err := getWorkflowOutput(succeeded, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)

		if err != nil {
			logger.Error(err, "could not retrieve workflow output, failing run configuration")
			newStatus.SynchronizationState = pipelinesv1.Failed
		} else {
			newStatus.KfpId = idResult
		}
	} else {
		if failed != nil {
			logger.Info("run configuration update failed")
		} else {
			logger.Info("run configuration creation progress unknown, failing run configuration")
		}
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
