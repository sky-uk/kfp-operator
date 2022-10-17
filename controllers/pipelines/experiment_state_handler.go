package pipelines

import (
	"context"
	"errors"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ExperimentStateHandler struct {
	WorkflowFactory    WorkflowFactory[*pipelinesv1.Experiment]
	WorkflowRepository WorkflowRepository
}

func (st *ExperimentStateHandler) stateTransition(ctx context.Context, experiment *pipelinesv1.Experiment) (commands []Command) {
	switch experiment.Status.SynchronizationState {
	case apis.Creating:
		commands = st.onCreating(ctx, experiment,
			st.WorkflowRepository.GetByLabels(ctx, experiment.GetNamespace(),
				CommonWorkflowLabels(experiment, WorkflowConstants.CreateOperationLabel)))
	case apis.Succeeded, apis.Failed:
		if !experiment.ObjectMeta.DeletionTimestamp.IsZero() {
			commands = st.onDelete(ctx, experiment)
		} else {
			commands = st.onSucceededOrFailed(ctx, experiment)
		}
	case apis.Updating:
		commands = st.onUpdating(ctx, experiment,
			st.WorkflowRepository.GetByLabels(ctx, experiment.GetNamespace(),
				CommonWorkflowLabels(experiment, WorkflowConstants.UpdateOperationLabel)))
	case apis.Deleting:
		commands = st.onDeleting(ctx, experiment,
			st.WorkflowRepository.GetByLabels(ctx, experiment.GetNamespace(),
				CommonWorkflowLabels(experiment, WorkflowConstants.DeleteOperationLabel)))
	case apis.Deleted:
	default:
		commands = st.onUnknown(ctx, experiment)
	}

	if experiment.Status.SynchronizationState == apis.Deleted {
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
		workflow, err := st.WorkflowFactory.ConstructUpdateWorkflow(experiment)

		if err != nil {
			failureMessage := WorkflowConstants.ConstructionFailedError
			logger.Error(err, fmt.Sprintf("%s, failing experiment", failureMessage))

			return []Command{
				*From(experiment.Status).WithSynchronizationState(apis.Failed).
					WithVersion(newExperimentVersion).
					WithMessage(failureMessage),
			}
		}

		return []Command{
			*From(experiment.Status).
				WithSynchronizationState(apis.Updating).
				WithVersion(newExperimentVersion),
			CreateWorkflow{Workflow: *workflow},
		}
	}

	logger.Info("empty state, creating experiment")
	workflow, err := st.WorkflowFactory.ConstructCreationWorkflow(experiment)

	if err != nil {
		failureMessage := WorkflowConstants.ConstructionFailedError
		logger.Error(err, fmt.Sprintf("%s, failing experiment", failureMessage))

		return []Command{
			*From(experiment.Status).WithSynchronizationState(apis.Failed).
				WithVersion(newExperimentVersion).
				WithMessage(failureMessage),
		}
	}

	return []Command{
		SetStatus{
			Status: apis.Status{
				Version:              newExperimentVersion,
				SynchronizationState: apis.Creating,
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
			*From(experiment.Status).WithSynchronizationState(apis.Deleted),
		}
	}

	workflow, err := st.WorkflowFactory.ConstructDeletionWorkflow(experiment)

	if err != nil {
		failureMessage := WorkflowConstants.ConstructionFailedError
		logger.Error(err, fmt.Sprintf("%s, failing experiment", failureMessage))

		return []Command{
			*From(experiment.Status).WithSynchronizationState(apis.Failed).WithMessage(failureMessage),
		}
	}

	return []Command{
		*From(experiment.Status).WithSynchronizationState(apis.Deleting),
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
	var targetState apis.SynchronizationState

	if experiment.Status.KfpId == "" {
		logger.Info("no kfpId exists, creating")
		workflow, err = st.WorkflowFactory.ConstructCreationWorkflow(experiment)

		if err != nil {
			failureMessage := WorkflowConstants.ConstructionFailedError
			logger.Error(err, fmt.Sprintf("%s, failing experiment", failureMessage))

			return []Command{
				*From(experiment.Status).
					WithSynchronizationState(apis.Failed).
					WithVersion(newExperimentVersion).
					WithMessage(failureMessage),
			}
		}

		targetState = apis.Creating
	} else {
		logger.Info("kfpId exists, updating")
		workflow, err = st.WorkflowFactory.ConstructUpdateWorkflow(experiment)

		if err != nil {
			failureMessage := WorkflowConstants.ConstructionFailedError
			logger.Error(err, fmt.Sprintf("%s, failing experiment", failureMessage))

			return []Command{
				*From(experiment.Status).
					WithSynchronizationState(apis.Failed).
					WithVersion(newExperimentVersion).
					WithMessage(failureMessage),
			}
		}

		targetState = apis.Updating
	}

	return []Command{
		*From(experiment.Status).
			WithSynchronizationState(targetState).
			WithVersion(newExperimentVersion),
		CreateWorkflow{Workflow: *workflow},
	}
}

type IdVerifier struct {
	SuccessState apis.SynchronizationState
	FailureState apis.SynchronizationState
	VerifyId     func(string) error
}

var succeedForEmptyId = IdVerifier{
	SuccessState: apis.Succeeded,
	FailureState: apis.Failed,
	VerifyId: func(id string) error {
		if id == "" {
			return errors.New("id was empty")
		}

		return nil
	},
}

var deletedForNonEmptyId = IdVerifier{
	SuccessState: apis.Deleted,
	FailureState: apis.Deleting,
	VerifyId: func(id string) error {
		if id != "" {
			return errors.New("id should be empty")
		}

		return nil
	},
}

func (st ExperimentStateHandler) setStateIfProviderFinished(ctx context.Context, status apis.Status, workflows []argo.Workflow, states IdVerifier) []Command {
	logger := log.FromContext(ctx)

	statusFromProviderOutput := func(workflow *argo.Workflow) *SetStatus {

		result, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)

		if err != nil {
			failureMessage := "could not retrieve workflow output"
			logger.Error(err, fmt.Sprintf("%s, failing experiment", failureMessage))
			return From(status).WithSynchronizationState(states.FailureState).WithMessage(failureMessage)
		}

		if result.ProviderError != "" {
			logger.Error(err, fmt.Sprintf("%s, failing experiment", result.ProviderError))
			return From(status).WithSynchronizationState(states.FailureState).WithMessage(result.ProviderError).WithKfpId(result.Id)
		}

		err = states.VerifyId(result.Id)

		if err != nil {
			failureMessage := err.Error()
			logger.Error(err, fmt.Sprintf("%s, failing experiment", failureMessage))
			return From(status).WithSynchronizationState(states.FailureState).WithMessage(failureMessage)
		}

		return From(status).WithSynchronizationState(states.SuccessState).WithKfpId(result.Id)
	}

	inProgress, succeeded, failed := latestWorkflowByPhase(workflows)

	if inProgress != nil {
		logger.V(2).Info("operation in progress")
		return []Command{}
	}

	var setStatusCommand *SetStatus

	if succeeded != nil {
		logger.Info("operation succeeded")
		setStatusCommand = statusFromProviderOutput(succeeded)
	} else {
		var failureMessage string

		if failed != nil {
			failureMessage = "operation failed"
		} else {
			failureMessage = "operation progress unknown"
		}

		logger.Info(fmt.Sprintf("%s, failing experiment", failureMessage))
		setStatusCommand = From(status).WithSynchronizationState(states.FailureState).WithMessage(failureMessage)
	}

	return []Command{
		*setStatusCommand,
		DeleteWorkflows{
			Workflows: workflows,
		},
	}
}

func (st ExperimentStateHandler) onCreating(ctx context.Context, experiment *pipelinesv1.Experiment, creationWorkflows []argo.Workflow) []Command {
	logger := log.FromContext(ctx)

	if experiment.Status.Version == "" {
		failureMessage := "creating experiment with empty version"
		logger.Info(fmt.Sprintf("%s, failing experiment", failureMessage))

		return []Command{
			*From(experiment.Status).WithSynchronizationState(apis.Failed).WithMessage(failureMessage),
		}
	}

	return st.setStateIfProviderFinished(ctx, experiment.Status, creationWorkflows, succeedForEmptyId)
}

func (st ExperimentStateHandler) onUpdating(ctx context.Context, experiment *pipelinesv1.Experiment, updateWorkflows []argo.Workflow) []Command {
	logger := log.FromContext(ctx)

	if experiment.Status.Version == "" || experiment.Status.KfpId == "" {
		failureMessage := "updating experiment with empty version or kfpId"
		logger.Info(fmt.Sprintf("%s, failing experiment", failureMessage))

		return []Command{
			*From(experiment.Status).WithSynchronizationState(apis.Failed).WithMessage(failureMessage),
		}
	}

	return st.setStateIfProviderFinished(ctx, experiment.Status, updateWorkflows, succeedForEmptyId)
}

func (st ExperimentStateHandler) onDeleting(ctx context.Context, experiment *pipelinesv1.Experiment, deletionWorkflows []argo.Workflow) []Command {
	return st.setStateIfProviderFinished(ctx, experiment.Status, deletionWorkflows, deletedForNonEmptyId)
}
