package pipelines

import (
	"context"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type RunConfigurationStateHandler struct {
	WorkflowFactory    RunConfigurationWorkflowFactory
	WorkflowRepository WorkflowRepository
}

func (st *RunConfigurationStateHandler) StateTransition(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) (commands []Command) {
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
		commands = append([]Command{ReleaseResource{}}, commands...)
	} else {
		commands = append([]Command{AcquireResource{}}, commands...)
	}

	return
}

func (st *RunConfigurationStateHandler) onUnknown(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) []Command {
	logger := log.FromContext(ctx)

	newRunConfigurationVErsion := runConfiguration.Spec.ComputeVersion()

	if runConfiguration.Status.KfpId != "" {
		logger.Info("empty state but KfpId already exists, updating runConfiguration")
		workflow, err := st.WorkflowFactory.ConstructUpdateWorkflow(ctx, runConfiguration)

		if err != nil {
			failureMessage := "error constructing update workflow"
			logger.Error(err, fmt.Sprintf("%s, failing run configuration", failureMessage))

			return []Command{
				*From(runConfiguration.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage),
			}
		}

		return []Command{
			*From(runConfiguration.Status).
				WithSynchronizationState(pipelinesv1.Updating).
				WithVersion(newRunConfigurationVErsion),
			CreateWorkflow{Workflow: *workflow},
		}
	}

	logger.Info("empty state, creating runConfiguration")
	workflow, err := st.WorkflowFactory.ConstructCreationWorkflow(ctx, runConfiguration)

	if err != nil {
		failureMessage := "error constructing creation workflow"
		logger.Error(err, fmt.Sprintf("%s, failing run configuration", failureMessage))

		return []Command{
			*From(runConfiguration.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage),
		}
	}

	return []Command{
		SetStatus{
			Status: pipelinesv1.Status{
				Version:              newRunConfigurationVErsion,
				SynchronizationState: pipelinesv1.Creating,
			},
		},
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st RunConfigurationStateHandler) onDelete(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) []Command {
	logger := log.FromContext(ctx)
	logger.Info("deletion requested, deleting")

	if runConfiguration.Status.KfpId == "" {
		return []Command{
			*From(runConfiguration.Status).WithSynchronizationState(pipelinesv1.Deleted),
		}
	}

	workflow, err := st.WorkflowFactory.ConstructDeletionWorkflow(ctx, runConfiguration)

	if err != nil {
		failureMessage := "error constructing deletion workflow"
		logger.Error(err, fmt.Sprintf("%s, failing run configuration", failureMessage))

		return []Command{
			*From(runConfiguration.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage),
		}
	}

	return []Command{
		*From(runConfiguration.Status).WithSynchronizationState(pipelinesv1.Deleting),
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st RunConfigurationStateHandler) onSucceededOrFailed(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) []Command {
	logger := log.FromContext(ctx)
	newRunConfigurationVersion := runConfiguration.Spec.ComputeVersion()

	if runConfiguration.Status.Version == newRunConfigurationVersion {
		logger.V(2).Info("run configuration version has not changed")
		return []Command{}
	}

	var workflow *argo.Workflow
	var err error
	var targetState pipelinesv1.SynchronizationState

	if runConfiguration.Status.KfpId == "" {
		logger.Info("no kfpId exists, creating")
		workflow, err = st.WorkflowFactory.ConstructCreationWorkflow(ctx, runConfiguration)

		if err != nil {
			failureMessage := "error constructing creation workflow"
			logger.Error(err, fmt.Sprintf("%s, failing run configuration", failureMessage))

			return []Command{
				*From(runConfiguration.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage),
			}
		}

		targetState = pipelinesv1.Creating
	} else {
		logger.Info("kfpId exists, updating")
		workflow, err = st.WorkflowFactory.ConstructUpdateWorkflow(ctx, runConfiguration)

		if err != nil {
			failureMessage := "error constructing update workflow"
			logger.Error(err, fmt.Sprintf("%s, failing run configuration", failureMessage))

			return []Command{
				*From(runConfiguration.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage),
			}
		}

		targetState = pipelinesv1.Updating
	}

	return []Command{
		*From(runConfiguration.Status).
			WithSynchronizationState(targetState).
			WithVersion(newRunConfigurationVersion),
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st RunConfigurationStateHandler) onUpdating(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration, updateWorkflows []argo.Workflow) []Command {
	logger := log.FromContext(ctx)

	if runConfiguration.Status.Version == "" || runConfiguration.Status.KfpId == "" {
		failureMessage := "updating run configuration with empty version or kfpId"
		logger.Info(fmt.Sprintf("%s, failing run configuration", failureMessage))

		return []Command{
			*From(runConfiguration.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage),
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
			return From(runConfiguration.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage)
		}

		idResult, _ := getWorkflowOutput(succeeded, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)

		if idResult == "" {
			failureMessage := "could not retrieve kfpId"
			logger.Info(fmt.Sprintf("%s, failing run configuration", failureMessage))
			return From(runConfiguration.Status).WithSynchronizationState(pipelinesv1.Failed).WithKfpId("").WithMessage(failureMessage)
		}

		logger.Info("run configuration update succeeded")
		return From(runConfiguration.Status).WithSynchronizationState(pipelinesv1.Succeeded).WithKfpId(idResult)
	}

	return []Command{
		*statusAfterUpdating(),
		DeleteWorkflows{
			Workflows: updateWorkflows,
		},
	}
}

func (st RunConfigurationStateHandler) onDeleting(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration, deletionWorkflows []argo.Workflow) (commands []Command) {
	logger := log.FromContext(ctx)

	inProgress, succeeded, failed := latestWorkflowByPhase(deletionWorkflows)

	if inProgress != nil {
		logger.V(2).Info("run configuration deletion in progress")
		return
	}

	var setStatusCommand *SetStatus

	if succeeded != nil {
		logger.Info("run configuration deletion succeeded")
		setStatusCommand = From(runConfiguration.Status).WithSynchronizationState(pipelinesv1.Deleted)
	} else {
		var failureMessage string

		if failed != nil {
			failureMessage = "run configuration deletion failed"
		} else {
			failureMessage = "run configuration deletion progress unknown"
		}

		logger.Info(failureMessage)
		setStatusCommand = From(runConfiguration.Status).WithMessage(failureMessage)
	}

	return []Command{
		*setStatusCommand,
		DeleteWorkflows{
			Workflows: deletionWorkflows,
		},
	}
}

func (st RunConfigurationStateHandler) onCreating(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration, creationWorkflows []argo.Workflow) (commands []Command) {
	logger := log.FromContext(ctx)

	if runConfiguration.Status.Version == "" {
		failureMessage := "creating run configuration with empty version"
		logger.Info(fmt.Sprintf("%s, failing run configuration", failureMessage))

		return []Command{
			*From(runConfiguration.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage),
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
			setStatusCommand = From(runConfiguration.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage)
		} else {
			setStatusCommand = From(runConfiguration.Status).WithSynchronizationState(pipelinesv1.Succeeded).WithKfpId(idResult)
		}
	} else {
		var failureMessage string

		if failed != nil {
			failureMessage = "run configuration creation failed"
		} else {
			failureMessage = "run configuration creation progress unknown"
		}

		logger.Info(fmt.Sprintf("%s, failing run configuration", failureMessage))
		setStatusCommand = From(runConfiguration.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage)
	}

	return []Command{
		*setStatusCommand,
		DeleteWorkflows{
			Workflows: creationWorkflows,
		},
	}
}
