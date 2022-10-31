package pipelines

import (
	"context"
	"errors"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type StateHandler[R pipelinesv1.Resource] struct {
	WorkflowFactory    WorkflowFactory[R]
	WorkflowRepository WorkflowRepository
}

func (st *StateHandler[R]) stateTransition(ctx context.Context, resource R) (commands []Command) {
	switch resource.GetStatus().SynchronizationState {
	case apis.Creating:
		commands = st.onCreating(ctx, resource,
			st.WorkflowRepository.GetByLabels(ctx, resource.GetNamespace(),
				CommonWorkflowLabels(resource, WorkflowConstants.CreateOperationLabel)))
	case apis.Succeeded, apis.Failed:
		if !resource.GetDeletionTimestamp().IsZero() {
			commands = st.onDelete(ctx, resource)
		} else {
			commands = st.onSucceededOrFailed(ctx, resource)
		}
	case apis.Updating:
		commands = st.onUpdating(ctx, resource,
			st.WorkflowRepository.GetByLabels(ctx, resource.GetNamespace(),
				CommonWorkflowLabels(resource, WorkflowConstants.UpdateOperationLabel)))
	case apis.Deleting:
		commands = st.onDeleting(ctx, resource,
			st.WorkflowRepository.GetByLabels(ctx, resource.GetNamespace(),
				CommonWorkflowLabels(resource, WorkflowConstants.DeleteOperationLabel)))
	case apis.Deleted:
	default:
		commands = st.onUnknown(ctx, resource)
	}

	if resource.GetStatus().SynchronizationState == apis.Deleted {
		commands = append([]Command{ReleaseResource{}}, commands...)
	} else {
		commands = append([]Command{AcquireResource{}}, commands...)
	}

	return
}

func (st *StateHandler[R]) StateTransition(ctx context.Context, resource R) []Command {
	logger := log.FromContext(ctx)
	logger.Info("state transition start")

	stateTransitionCommands := st.stateTransition(ctx, resource)
	return alwaysSetObservedGeneration(ctx, stateTransitionCommands, resource)
}

func (st *StateHandler[R]) onUnknown(ctx context.Context, resource R) []Command {
	logger := log.FromContext(ctx)

	newExperimentVersion := resource.ComputeVersion()

	if resource.GetStatus().ProviderId != "" {
		logger.Info("empty state but ProviderId already exists, updating resource")
		workflow, err := st.WorkflowFactory.ConstructUpdateWorkflow(resource)

		if err != nil {
			failureMessage := WorkflowConstants.ConstructionFailedError
			logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))

			return []Command{
				*From(resource.GetStatus()).WithSynchronizationState(apis.Failed).
					WithVersion(newExperimentVersion).
					WithMessage(failureMessage),
			}
		}

		return []Command{
			*From(resource.GetStatus()).
				WithSynchronizationState(apis.Updating).
				WithVersion(newExperimentVersion),
			CreateWorkflow{Workflow: *workflow},
		}
	}

	logger.Info("empty state, creating resource")
	workflow, err := st.WorkflowFactory.ConstructCreationWorkflow(resource)

	if err != nil {
		failureMessage := WorkflowConstants.ConstructionFailedError
		logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))

		return []Command{
			*From(resource.GetStatus()).WithSynchronizationState(apis.Failed).
				WithVersion(newExperimentVersion).
				WithMessage(failureMessage),
		}
	}

	return []Command{
		SetStatus{
			Status: pipelinesv1.Status{
				Version:              newExperimentVersion,
				SynchronizationState: apis.Creating,
			},
		},
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st StateHandler[R]) onDelete(ctx context.Context, resource R) []Command {
	logger := log.FromContext(ctx)
	logger.Info("deletion requested, deleting")

	if resource.GetStatus().ProviderId == "" {
		return []Command{
			*From(resource.GetStatus()).WithSynchronizationState(apis.Deleted),
		}
	}

	workflow, err := st.WorkflowFactory.ConstructDeletionWorkflow(resource)

	if err != nil {
		failureMessage := WorkflowConstants.ConstructionFailedError
		logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))

		return []Command{
			*From(resource.GetStatus()).WithSynchronizationState(apis.Failed).WithMessage(failureMessage),
		}
	}

	return []Command{
		*From(resource.GetStatus()).WithSynchronizationState(apis.Deleting),
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st StateHandler[R]) onSucceededOrFailed(ctx context.Context, resource R) []Command {
	logger := log.FromContext(ctx)
	newExperimentVersion := resource.ComputeVersion()

	if resource.GetStatus().Version == newExperimentVersion {
		logger.V(2).Info("resource version has not changed")
		return []Command{}
	}

	var workflow *argo.Workflow
	var err error
	var targetState apis.SynchronizationState

	if resource.GetStatus().ProviderId == "" {
		logger.Info("no providerId exists, creating")
		workflow, err = st.WorkflowFactory.ConstructCreationWorkflow(resource)

		if err != nil {
			failureMessage := WorkflowConstants.ConstructionFailedError
			logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))

			return []Command{
				*From(resource.GetStatus()).
					WithSynchronizationState(apis.Failed).
					WithVersion(newExperimentVersion).
					WithMessage(failureMessage),
			}
		}

		targetState = apis.Creating
	} else {
		logger.Info("providerId exists, updating")
		workflow, err = st.WorkflowFactory.ConstructUpdateWorkflow(resource)

		if err != nil {
			failureMessage := WorkflowConstants.ConstructionFailedError
			logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))

			return []Command{
				*From(resource.GetStatus()).
					WithSynchronizationState(apis.Failed).
					WithVersion(newExperimentVersion).
					WithMessage(failureMessage),
			}
		}

		targetState = apis.Updating
	}

	return []Command{
		*From(resource.GetStatus()).
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

func (st StateHandler[R]) setStateIfProviderFinished(ctx context.Context, status pipelinesv1.Status, workflows []argo.Workflow, states IdVerifier) []Command {
	logger := log.FromContext(ctx)

	statusFromProviderOutput := func(workflow *argo.Workflow) *SetStatus {

		result, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)

		if err != nil {
			failureMessage := "could not retrieve workflow output"
			logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))
			return From(status).WithSynchronizationState(states.FailureState).WithMessage(failureMessage)
		}

		if result.ProviderError != "" {
			logger.Error(err, fmt.Sprintf("%s, failing resource", result.ProviderError))
			return From(status).WithSynchronizationState(states.FailureState).WithMessage(result.ProviderError).WithProviderId(result.Id)
		}

		err = states.VerifyId(result.Id)

		if err != nil {
			failureMessage := err.Error()
			logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))
			return From(status).WithSynchronizationState(states.FailureState).WithMessage(failureMessage)
		}

		return From(status).WithSynchronizationState(states.SuccessState).WithProviderId(result.Id)
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

		logger.Info(fmt.Sprintf("%s, failing resource", failureMessage))
		setStatusCommand = From(status).WithSynchronizationState(states.FailureState).WithMessage(failureMessage)
	}

	return []Command{
		*setStatusCommand,
		DeleteWorkflows{
			Workflows: workflows,
		},
	}
}

func (st StateHandler[R]) onCreating(ctx context.Context, resource R, creationWorkflows []argo.Workflow) []Command {
	logger := log.FromContext(ctx)

	if resource.GetStatus().Version == "" {
		failureMessage := "creating resource with empty version"
		logger.Info(fmt.Sprintf("%s, failing resource", failureMessage))

		return []Command{
			*From(resource.GetStatus()).WithSynchronizationState(apis.Failed).WithMessage(failureMessage),
		}
	}

	return st.setStateIfProviderFinished(ctx, resource.GetStatus(), creationWorkflows, succeedForEmptyId)
}

func (st StateHandler[R]) onUpdating(ctx context.Context, resource R, updateWorkflows []argo.Workflow) []Command {
	logger := log.FromContext(ctx)

	if resource.GetStatus().Version == "" || resource.GetStatus().ProviderId == "" {
		failureMessage := "updating resource with empty version or providerId"
		logger.Info(fmt.Sprintf("%s, failing resource", failureMessage))

		return []Command{
			*From(resource.GetStatus()).WithSynchronizationState(apis.Failed).WithMessage(failureMessage),
		}
	}

	return st.setStateIfProviderFinished(ctx, resource.GetStatus(), updateWorkflows, succeedForEmptyId)
}

func (st StateHandler[R]) onDeleting(ctx context.Context, resource R, deletionWorkflows []argo.Workflow) []Command {
	return st.setStateIfProviderFinished(ctx, resource.GetStatus(), deletionWorkflows, deletedForNonEmptyId)
}
