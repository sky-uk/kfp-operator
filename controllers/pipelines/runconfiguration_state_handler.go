package pipelines

import (
	"context"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type RunConfigurationStateHandler struct {
	WorkflowFactory    WorkflowFactory[*pipelinesv1.RunConfiguration]
	WorkflowRepository WorkflowRepository
}

func (st *RunConfigurationStateHandler) stateTransition(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) (commands []Command) {
	switch runConfiguration.Status.Status.SynchronizationState {
	case apis.Creating:
		commands = st.onCreating(ctx, runConfiguration,
			st.WorkflowRepository.GetByLabels(ctx, runConfiguration.GetNamespace(),
				CommonWorkflowLabels(runConfiguration, WorkflowConstants.CreateOperationLabel)))
	case apis.Succeeded, apis.Failed:
		if !runConfiguration.ObjectMeta.DeletionTimestamp.IsZero() {
			commands = st.onDelete(ctx, runConfiguration)
		} else {
			commands = st.onSucceededOrFailed(ctx, runConfiguration)
		}
	case apis.Updating:
		commands = st.onUpdating(ctx, runConfiguration,
			st.WorkflowRepository.GetByLabels(ctx, runConfiguration.GetNamespace(),
				CommonWorkflowLabels(runConfiguration, WorkflowConstants.UpdateOperationLabel)))
	case apis.Deleting:
		commands = st.onDeleting(ctx, runConfiguration,
			st.WorkflowRepository.GetByLabels(ctx, runConfiguration.GetNamespace(),
				CommonWorkflowLabels(runConfiguration, WorkflowConstants.DeleteOperationLabel)))
	case apis.Deleted:
	default:
		commands = st.onUnknown(ctx, runConfiguration)
	}

	if runConfiguration.Status.Status.SynchronizationState == apis.Deleted {
		commands = append([]Command{ReleaseResource{}}, commands...)
	} else {
		commands = append([]Command{AcquireResource{}}, commands...)
	}

	return
}

func (st *RunConfigurationStateHandler) StateTransition(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) []Command {
	logger := log.FromContext(ctx)
	logger.Info("state transition start")

	stateTransitionCommands := st.stateTransition(ctx, runConfiguration)
	return alwaysSetObservedGeneration(ctx, stateTransitionCommands, runConfiguration)
}

func (st *RunConfigurationStateHandler) onUnknown(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) []Command {
	logger := log.FromContext(ctx)

	newRunConfigurationVersion := runConfiguration.ComputeVersion()

	if runConfiguration.Status.Status.KfpId != "" {
		logger.Info("empty state but KfpId already exists, updating runConfiguration")
		workflow, err := st.WorkflowFactory.ConstructUpdateWorkflow(runConfiguration)

		if err != nil {
			failureMessage := WorkflowConstants.ConstructionFailedError
			logger.Error(err, fmt.Sprintf("%s, failing run configuration", failureMessage))

			return []Command{
				*From(runConfiguration.Status.Status).
					WithSynchronizationState(apis.Failed).
					WithVersion(newRunConfigurationVersion).
					WithMessage(failureMessage),
			}
		}

		return []Command{
			*From(runConfiguration.Status.Status).
				WithSynchronizationState(apis.Updating).
				WithVersion(newRunConfigurationVersion),
			CreateWorkflow{Workflow: *workflow},
		}
	}

	logger.Info("empty state, creating runConfiguration")
	workflow, err := st.WorkflowFactory.ConstructCreationWorkflow(runConfiguration)

	if err != nil {
		failureMessage := WorkflowConstants.ConstructionFailedError
		logger.Error(err, fmt.Sprintf("%s, failing run configuration", failureMessage))

		return []Command{
			*From(runConfiguration.Status.Status).
				WithSynchronizationState(apis.Failed).
				WithVersion(newRunConfigurationVersion).
				WithMessage(failureMessage),
		}
	}

	return []Command{
		SetStatus{
			Status: apis.Status{
				Version:              newRunConfigurationVersion,
				SynchronizationState: apis.Creating,
			},
		},
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st RunConfigurationStateHandler) onDelete(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) []Command {
	logger := log.FromContext(ctx)
	logger.Info("deletion requested, deleting")

	if runConfiguration.Status.Status.KfpId == "" {
		return []Command{
			*From(runConfiguration.Status.Status).WithSynchronizationState(apis.Deleted),
		}
	}

	workflow, err := st.WorkflowFactory.ConstructDeletionWorkflow(runConfiguration)

	if err != nil {
		failureMessage := WorkflowConstants.ConstructionFailedError
		logger.Error(err, fmt.Sprintf("%s, failing run configuration", failureMessage))

		return []Command{
			*From(runConfiguration.Status.Status).WithSynchronizationState(apis.Failed).WithMessage(failureMessage),
		}
	}

	return []Command{
		*From(runConfiguration.Status.Status).WithSynchronizationState(apis.Deleting),
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st RunConfigurationStateHandler) onSucceededOrFailed(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) []Command {
	logger := log.FromContext(ctx)
	newRunConfigurationVersion := runConfiguration.ComputeVersion()

	if runConfiguration.Status.Status.Version == newRunConfigurationVersion {
		logger.V(2).Info("run configuration version has not changed")
		return []Command{}
	}

	var workflow *argo.Workflow
	var err error
	var targetState apis.SynchronizationState

	if runConfiguration.Status.Status.KfpId == "" {
		logger.Info("no kfpId exists, creating")
		workflow, err = st.WorkflowFactory.ConstructCreationWorkflow(runConfiguration)

		if err != nil {
			failureMessage := WorkflowConstants.ConstructionFailedError
			logger.Error(err, fmt.Sprintf("%s, failing run configuration", failureMessage))

			return []Command{
				*From(runConfiguration.Status.Status).
					WithSynchronizationState(apis.Failed).
					WithVersion(newRunConfigurationVersion).
					WithMessage(failureMessage),
			}
		}

		targetState = apis.Creating
	} else {
		logger.Info("kfpId exists, updating")
		workflow, err = st.WorkflowFactory.ConstructUpdateWorkflow(runConfiguration)

		if err != nil {
			failureMessage := WorkflowConstants.ConstructionFailedError
			logger.Error(err, fmt.Sprintf("%s, failing run configuration", failureMessage))

			return []Command{
				*From(runConfiguration.Status.Status).
					WithSynchronizationState(apis.Failed).
					WithVersion(newRunConfigurationVersion).
					WithMessage(failureMessage),
			}
		}

		targetState = apis.Updating
	}

	return []Command{
		*From(runConfiguration.Status.Status).
			WithSynchronizationState(targetState).
			WithVersion(newRunConfigurationVersion),
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st RunConfigurationStateHandler) onUpdating(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration, updateWorkflows []argo.Workflow) []Command {
	logger := log.FromContext(ctx)

	if runConfiguration.Status.Status.Version == "" || runConfiguration.Status.Status.KfpId == "" {
		failureMessage := "updating run configuration with empty version or kfpId"
		logger.Info(fmt.Sprintf("%s, failing run configuration", failureMessage))

		return []Command{
			*From(runConfiguration.Status.Status).WithSynchronizationState(apis.Failed).WithMessage(failureMessage),
		}
	}

	inProgress, succeeded, failed := latestWorkflowByPhase(updateWorkflows)

	if inProgress != nil {
		logger.V(2).Info("run configuration update in progress")
		return []Command{}
	}

	statusAfterUpdating := func() *SetStatus {
		if succeeded == nil {
			var failureMessage string

			if failed != nil {
				failureMessage = "run configuration update failed"
			} else {
				failureMessage = "run configuration creation progress unknown"
			}

			logger.Info(fmt.Sprintf("%s, failing run configuration", failureMessage))
			return From(runConfiguration.Status.Status).WithSynchronizationState(apis.Failed).WithMessage(failureMessage)
		}

		idResult, _ := getWorkflowOutput(succeeded, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)

		if idResult == "" {
			failureMessage := "could not retrieve kfpId"
			logger.Info(fmt.Sprintf("%s, failing run configuration", failureMessage))
			return From(runConfiguration.Status.Status).WithSynchronizationState(apis.Failed).WithKfpId("").WithMessage(failureMessage)
		}

		logger.Info("run configuration update succeeded")
		return From(runConfiguration.Status.Status).WithSynchronizationState(apis.Succeeded).WithKfpId(idResult)
	}

	return []Command{
		*statusAfterUpdating(),
		DeleteWorkflows{
			Workflows: updateWorkflows,
		},
	}
}

func (st RunConfigurationStateHandler) onDeleting(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration, deletionWorkflows []argo.Workflow) []Command {
	logger := log.FromContext(ctx)

	inProgress, succeeded, failed := latestWorkflowByPhase(deletionWorkflows)

	if inProgress != nil {
		logger.V(2).Info("run configuration deletion in progress")
		return []Command{}
	}

	var setStatusCommand *SetStatus

	if succeeded != nil {
		logger.Info("run configuration deletion succeeded")
		setStatusCommand = From(runConfiguration.Status.Status).WithSynchronizationState(apis.Deleted)
	} else {
		var failureMessage string

		if failed != nil {
			failureMessage = "run configuration deletion failed"
		} else {
			failureMessage = "run configuration deletion progress unknown"
		}

		logger.Info(failureMessage)
		setStatusCommand = From(runConfiguration.Status.Status).WithMessage(failureMessage)
	}

	return []Command{
		*setStatusCommand,
		DeleteWorkflows{
			Workflows: deletionWorkflows,
		},
	}
}

func (st RunConfigurationStateHandler) onCreating(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration, creationWorkflows []argo.Workflow) []Command {
	logger := log.FromContext(ctx)

	if runConfiguration.Status.Status.Version == "" {
		failureMessage := "creating run configuration with empty version"
		logger.Info(fmt.Sprintf("%s, failing run configuration", failureMessage))

		return []Command{
			*From(runConfiguration.Status.Status).WithSynchronizationState(apis.Failed).WithMessage(failureMessage),
		}
	}

	inProgress, succeeded, failed := latestWorkflowByPhase(creationWorkflows)

	if inProgress != nil {
		logger.V(2).Info("run configuration creation in progress")
		return []Command{}
	}

	var setStatusCommand *SetStatus

	if succeeded != nil {
		logger.Info("run configuration creation succeeded")
		idResult, err := getWorkflowOutput(succeeded, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)

		if err != nil {
			failureMessage := "could not retrieve workflow output"
			logger.Error(err, fmt.Sprintf("%s, failing run configuration", failureMessage))
			setStatusCommand = From(runConfiguration.Status.Status).WithSynchronizationState(apis.Failed).WithMessage(failureMessage)
		} else {
			setStatusCommand = From(runConfiguration.Status.Status).WithSynchronizationState(apis.Succeeded).WithKfpId(idResult)
		}
	} else {
		var failureMessage string

		if failed != nil {
			failureMessage = "run configuration creation failed"
		} else {
			failureMessage = "run configuration creation progress unknown"
		}

		logger.Info(fmt.Sprintf("%s, failing run configuration", failureMessage))
		setStatusCommand = From(runConfiguration.Status.Status).WithSynchronizationState(apis.Failed).WithMessage(failureMessage)
	}

	return []Command{
		*setStatusCommand,
		DeleteWorkflows{
			Workflows: creationWorkflows,
		},
	}
}
