package pipelines

import (
	"context"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ExperimentStateHandler struct {
	WorkflowFactory    ExperimentWorkflowFactory
	WorkflowRepository WorkflowRepository
}

func (st *ExperimentStateHandler) stateTransition(ctx context.Context, experiment *pipelinesv1.Experiment) (commands []Command) {
	switch experiment.Status.SynchronizationState {
	case pipelinesv1.Creating:
		commands = st.onCreating(ctx, experiment,
			st.WorkflowRepository.GetByLabels(ctx, experiment.NamespacedName(), map[string]string{
				PipelineWorkflowConstants.OperationLabelKey: ExperimentWorkflowConstants.CreateOperationLabel,
				ExperimentWorkflowConstants.ExperimentNameLabelKey: experiment.GetName(),
			}))
	case pipelinesv1.Succeeded, pipelinesv1.Failed:
		if !experiment.ObjectMeta.DeletionTimestamp.IsZero() {
			commands = st.onDelete(ctx, experiment)
		} else {
			commands = st.onSucceededOrFailed(ctx, experiment)
		}
	case pipelinesv1.Updating:
		commands = st.onUpdating(ctx, experiment,
			st.WorkflowRepository.GetByLabels(ctx, experiment.NamespacedName(), map[string]string{
				PipelineWorkflowConstants.OperationLabelKey: ExperimentWorkflowConstants.UpdateOperationLabel,
				ExperimentWorkflowConstants.ExperimentNameLabelKey: experiment.GetName(),
			}))
	case pipelinesv1.Deleting:
		commands = st.onDeleting(ctx, experiment,
			st.WorkflowRepository.GetByLabels(ctx, experiment.NamespacedName(), map[string]string{
				PipelineWorkflowConstants.OperationLabelKey: ExperimentWorkflowConstants.DeleteOperationLabel,
				ExperimentWorkflowConstants.ExperimentNameLabelKey: experiment.GetName(),
			}))
	case pipelinesv1.Deleted:
	default:
		commands = st.onUnknown(ctx, experiment)
	}

	if experiment.Status.SynchronizationState == pipelinesv1.Deleted {
		commands = append([]Command{ReleaseResource{}}, commands...)
	} else {
		commands = append([]Command{AcquireResource{}}, commands...)
	}

	return
}

func (st *ExperimentStateHandler) StateTransition(ctx context.Context, experiment *pipelinesv1.Experiment) []Command {
	logger := log.FromContext(ctx)
	logger.Info("state transition start")

	stateTransitionCommands := st.stateTransition(ctx, experiment)
	return alwaysSetObservedGeneration(ctx, stateTransitionCommands, experiment)
}

func (st *ExperimentStateHandler) onUnknown(ctx context.Context, experiment *pipelinesv1.Experiment) []Command {
	logger := log.FromContext(ctx)

	newExperimentVersion := experiment.Spec.ComputeVersion()

	if experiment.Status.KfpId != "" {
		logger.Info("empty state but KfpId already exists, updating experiment")
		workflow, err := st.WorkflowFactory.ConstructUpdateWorkflow(ctx, experiment)

		if err != nil {
			failureMessage := "error constructing update workflow"
			logger.Error(err, fmt.Sprintf("%s, failing experiment", failureMessage))

			return []Command{
				*From(experiment.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage),
			}
		}

		return []Command{
			*From(experiment.Status).
				WithSynchronizationState(pipelinesv1.Updating).
				WithVersion(newExperimentVersion),
			CreateWorkflow{Workflow: *workflow},
		}
	}

	logger.Info("empty state, creating experiment")
	workflow, err := st.WorkflowFactory.ConstructCreationWorkflow(ctx, experiment)

	if err != nil {
		failureMessage := "error constructing creation workflow"
		logger.Error(err, fmt.Sprintf("%s, failing experiment", failureMessage))

		return []Command{
			*From(experiment.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage),
		}
	}

	return []Command{
		SetStatus{
			Status: pipelinesv1.Status{
				Version:              newExperimentVersion,
				SynchronizationState: pipelinesv1.Creating,
			},
		},
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st ExperimentStateHandler) onDelete(ctx context.Context, experiment *pipelinesv1.Experiment) []Command {
	logger := log.FromContext(ctx)
	logger.Info("deletion requested, deleting")

	if experiment.Status.KfpId == "" {
		return []Command{
			*From(experiment.Status).WithSynchronizationState(pipelinesv1.Deleted),
		}
	}

	workflow, err := st.WorkflowFactory.ConstructDeletionWorkflow(ctx, experiment)

	if err != nil {
		failureMessage := "error constructing deletion workflow"
		logger.Error(err, fmt.Sprintf("%s, failing experiment", failureMessage))

		return []Command{
			*From(experiment.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage),
		}
	}

	return []Command{
		*From(experiment.Status).WithSynchronizationState(pipelinesv1.Deleting),
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st ExperimentStateHandler) onSucceededOrFailed(ctx context.Context, experiment *pipelinesv1.Experiment) []Command {
	logger := log.FromContext(ctx)
	newExperimentVersion := experiment.Spec.ComputeVersion()

	if experiment.Status.Version == newExperimentVersion {
		logger.V(2).Info("experiment version has not changed")
		return []Command{}
	}

	var workflow *argo.Workflow
	var err error
	var targetState pipelinesv1.SynchronizationState

	if experiment.Status.KfpId == "" {
		logger.Info("no kfpId exists, creating")
		workflow, err = st.WorkflowFactory.ConstructCreationWorkflow(ctx, experiment)

		if err != nil {
			failureMessage := "error constructing creation workflow"
			logger.Error(err, fmt.Sprintf("%s, failing experiment", failureMessage))

			return []Command{
				*From(experiment.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage),
			}
		}

		targetState = pipelinesv1.Creating
	} else {
		logger.Info("kfpId exists, updating")
		workflow, err = st.WorkflowFactory.ConstructUpdateWorkflow(ctx, experiment)

		if err != nil {
			failureMessage := "error constructing update workflow"
			logger.Error(err, fmt.Sprintf("%s, failing experiment", failureMessage))

			return []Command{
				*From(experiment.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage),
			}
		}

		targetState = pipelinesv1.Updating
	}

	return []Command{
		*From(experiment.Status).
			WithSynchronizationState(targetState).
			WithVersion(newExperimentVersion),
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st ExperimentStateHandler) onUpdating(ctx context.Context, experiment *pipelinesv1.Experiment, updateWorkflows []argo.Workflow) []Command {
	logger := log.FromContext(ctx)

	if experiment.Status.Version == "" || experiment.Status.KfpId == "" {
		failureMessage := "updating experiment with empty version or kfpId"
		logger.Info(fmt.Sprintf("%s, failing experiment", failureMessage))

		return []Command{
			*From(experiment.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage),
		}
	}

	inProgress, succeeded, failed := latestWorkflowByPhase(updateWorkflows)

	if inProgress != nil {
		logger.V(2).Info("experiment update in progress")
		return []Command{}
	}

	statusAfterUpdating := func() *SetStatus {
		if succeeded == nil {
			var failureMessage string

			if failed != nil {
				failureMessage = "experiment update failed"
			} else {
				failureMessage = "experiment creation progress unknown"
			}

			logger.Info(fmt.Sprintf("%s, failing experiment", failureMessage))
			return From(experiment.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage)
		}

		idResult, _ := getWorkflowOutput(succeeded, ExperimentWorkflowConstants.ExperimentIdParameterName)

		if idResult == "" {
			failureMessage := "could not retrieve kfpId"
			logger.Info(fmt.Sprintf("%s, failing experiment", failureMessage))
			return From(experiment.Status).WithSynchronizationState(pipelinesv1.Failed).WithKfpId("").WithMessage(failureMessage)
		}

		logger.Info("experiment update succeeded")
		return From(experiment.Status).WithSynchronizationState(pipelinesv1.Succeeded).WithKfpId(idResult)
	}

	return []Command{
		*statusAfterUpdating(),
		DeleteWorkflows{
			Workflows: updateWorkflows,
		},
	}
}

func (st ExperimentStateHandler) onDeleting(ctx context.Context, experiment *pipelinesv1.Experiment, deletionWorkflows []argo.Workflow) []Command {
	logger := log.FromContext(ctx)

	inProgress, succeeded, failed := latestWorkflowByPhase(deletionWorkflows)

	if inProgress != nil {
		logger.V(2).Info("experiment deletion in progress")
		return []Command{}
	}

	var setStatusCommand *SetStatus

	if succeeded != nil {
		logger.Info("experiment deletion succeeded")
		setStatusCommand = From(experiment.Status).WithSynchronizationState(pipelinesv1.Deleted)
	} else {
		var failureMessage string

		if failed != nil {
			failureMessage = "experiment deletion failed"
		} else {
			failureMessage = "experiment deletion progress unknown"
		}

		logger.Info(failureMessage)
		setStatusCommand = From(experiment.Status).WithMessage(failureMessage)
	}

	return []Command{
		*setStatusCommand,
		DeleteWorkflows{
			Workflows: deletionWorkflows,
		},
	}
}

func (st ExperimentStateHandler) onCreating(ctx context.Context, experiment *pipelinesv1.Experiment, creationWorkflows []argo.Workflow) []Command {
	logger := log.FromContext(ctx)

	if experiment.Status.Version == "" {
		failureMessage := "creating experiment with empty version"
		logger.Info(fmt.Sprintf("%s, failing experiment", failureMessage))

		return []Command{
			*From(experiment.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage),
		}
	}

	inProgress, succeeded, failed := latestWorkflowByPhase(creationWorkflows)

	if inProgress != nil {
		logger.V(2).Info("experiment creation in progress")
		return []Command{}
	}

	var setStatusCommand *SetStatus

	if succeeded != nil {
		logger.Info("experiment creation succeeded")
		idResult, err := getWorkflowOutput(succeeded, ExperimentWorkflowConstants.ExperimentIdParameterName)

		if err != nil {
			failureMessage := "could not retrieve workflow output"
			logger.Error(err, fmt.Sprintf("%s, failing experiment", failureMessage))
			setStatusCommand = From(experiment.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage)
		} else {
			setStatusCommand = From(experiment.Status).WithSynchronizationState(pipelinesv1.Succeeded).WithKfpId(idResult)
		}
	} else {
		var failureMessage string

		if failed != nil {
			failureMessage = "experiment creation failed"
		} else {
			failureMessage = "experiment creation progress unknown"
		}

		logger.Info(fmt.Sprintf("%s, failing experiment", failureMessage))
		setStatusCommand = From(experiment.Status).WithSynchronizationState(pipelinesv1.Failed).WithMessage(failureMessage)
	}

	return []Command{
		*setStatusCommand,
		DeleteWorkflows{
			Workflows: creationWorkflows,
		},
	}
}
