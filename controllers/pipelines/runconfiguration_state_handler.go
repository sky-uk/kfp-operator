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

func (st *RunConfigurationStateHandler) StateTransition(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) (commands []RunConfigurationCommand) {
	logger := log.FromContext(ctx)
	logger.Info("state transition start")

	switch runConfiguration.Status.SynchronizationState {
	case pipelinesv1.Creating:
		commands = st.onCreating(ctx, runConfiguration,
			st.WorkflowRepository.GetByOperation(ctx,
				RunConfigurationWorkflowConstants.CreateOperationLabel,
				runConfiguration.NamespacedName(),
				RunConfigurationWorkflowConstants.RunConfigurationNameLabelKey))
	case pipelinesv1.Succeeded, pipelinesv1.Failed:
		if !runConfiguration.ObjectMeta.DeletionTimestamp.IsZero() {
			commands = st.onDelete(ctx, runConfiguration)
		} else {
			commands = st.onSucceededOrFailed(ctx, runConfiguration)
		}
	case pipelinesv1.Updating:
		commands = st.onUpdating(ctx, runConfiguration,
			st.WorkflowRepository.GetByOperation(ctx,
				RunConfigurationWorkflowConstants.UpdateOperationLabel,
				runConfiguration.NamespacedName(),
				RunConfigurationWorkflowConstants.RunConfigurationNameLabelKey))
	case pipelinesv1.Deleting:
		commands = st.onDeleting(ctx, runConfiguration,
			st.WorkflowRepository.GetByOperation(ctx,
				RunConfigurationWorkflowConstants.DeleteOperationLabel,
				runConfiguration.NamespacedName(),
				RunConfigurationWorkflowConstants.RunConfigurationNameLabelKey))
	case pipelinesv1.Deleted:
	default:
		commands = st.onUnknown(ctx, runConfiguration)
	}

	if runConfiguration.Status.SynchronizationState == pipelinesv1.Deleted {
		commands = append([]RunConfigurationCommand{ReleaseRunConfiguration{}}, commands...)
	} else {
		commands = append([]RunConfigurationCommand{AcquireRunConfiguration{}}, commands...)
	}

	return
}

func (st *RunConfigurationStateHandler) onUnknown(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) []RunConfigurationCommand {
	logger := log.FromContext(ctx)

	newRunconfigurationVersion := runConfiguration.Spec.ComputeVersion()

	if runConfiguration.Status.KfpId != "" {
		logger.Info("empty state but KfpId already exists, updating run configuration")
		workflow, err := st.WorkflowFactory.ConstructUpdateWorkflow(ctx, runConfiguration)

		if err != nil {
			logger.Error(err, "error constructing update workflow, failing run configuration")

			return []RunConfigurationCommand{
				SetRunConfigurationStatus{
					Status: pipelinesv1.Status{
						Version:              runConfiguration.Status.Version,
						SynchronizationState: pipelinesv1.Failed,
					},
				},
			}
		}

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
	workflow, err := st.WorkflowFactory.ConstructCreationWorkflow(ctx, runConfiguration)
	if err != nil {
		logger.Error(err, "error constructing creation workflow, failing run configuration")

		return []RunConfigurationCommand{
			SetRunConfigurationStatus{
				Status: pipelinesv1.Status{
					Version:              runConfiguration.Status.Version,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

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

	if runConfiguration.Status.KfpId == "" {
		return []RunConfigurationCommand{
			SetRunConfigurationStatus{
				Status: pipelinesv1.Status{
					KfpId:                runConfiguration.Status.KfpId,
					Version:              runConfiguration.Status.Version,
					SynchronizationState: pipelinesv1.Deleted,
				},
			},
		}
	}

	workflow, err := st.WorkflowFactory.ConstructDeletionWorkflow(ctx, runConfiguration)
	if err != nil {
		logger.Error(err, "error constructing deletion workflow, failing run configuration")

		return []RunConfigurationCommand{
			SetRunConfigurationStatus{
				Status: pipelinesv1.Status{
					Version:              runConfiguration.Status.Version,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

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
	var err error
	var targetState pipelinesv1.SynchronizationState

	if runConfiguration.Status.KfpId == "" {
		logger.Info("no kfpId exists, creating")
		workflow, err = st.WorkflowFactory.ConstructCreationWorkflow(ctx, runConfiguration)
		if err != nil {
			logger.Error(err, "error constructing creation workflow, failing run configuration")

			return []RunConfigurationCommand{
				SetRunConfigurationStatus{
					Status: pipelinesv1.Status{
						Version:              runConfiguration.Status.Version,
						SynchronizationState: pipelinesv1.Failed,
					},
				},
			}
		}

		targetState = pipelinesv1.Creating
	} else {
		logger.Info("kfpId exists, updating")
		workflow, err = st.WorkflowFactory.ConstructUpdateWorkflow(ctx, runConfiguration)
		if err != nil {
			logger.Error(err, "error constructing update workflow, failing run configuration")

			return []RunConfigurationCommand{
				SetRunConfigurationStatus{
					Status: pipelinesv1.Status{
						Version:              runConfiguration.Status.Version,
						SynchronizationState: pipelinesv1.Failed,
					},
				},
			}
		}

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

	statusAfterUpdating := func() (newStatus pipelinesv1.Status) {
		newStatus = *runConfiguration.Status.DeepCopy()

		if succeeded == nil {
			if failed != nil {
				logger.Info("run configuration update failed")
			} else {
				logger.Info("run configuration creation progress unknown, failing run configuration")
			}

			newStatus.SynchronizationState = pipelinesv1.Failed
			return
		}

		idResult, _ := getWorkflowOutput(succeeded, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)

		if idResult == "" {
			logger.Info("could not retrieve kfpId, failing run configuration")
			newStatus.KfpId = ""
			newStatus.SynchronizationState = pipelinesv1.Failed
			return
		}

		logger.Info("run configuration update succeeded")
		newStatus.KfpId = idResult
		newStatus.SynchronizationState = pipelinesv1.Succeeded
		return
	}

	return []RunConfigurationCommand{
		SetRunConfigurationStatus{
			Status: statusAfterUpdating(),
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
