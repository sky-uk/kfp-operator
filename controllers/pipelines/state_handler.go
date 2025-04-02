package pipelines

import (
	"context"
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowfactory"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowutil"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type StateHandler[R pipelineshub.Resource] struct {
	WorkflowFactory    workflowfactory.WorkflowFactory[R]
	WorkflowRepository WorkflowRepository
}

var StateHandlerConstants = struct {
	ProviderChangedError string
}{
	ProviderChangedError: "the provider has changed",
}

func (st *StateHandler[R]) stateTransition(
	ctx context.Context,
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
	transitionTime metav1.Time,
) (commands []Command) {
	resourceProvider := resource.GetStatus().Provider.Name
	if !resourceProvider.Empty() && resourceProvider != provider.GetCommonNamespacedName() {
		setStatus := From(resource.GetStatus()).
			WithMessage(StateHandlerConstants.ProviderChangedError).
			WithLastTransitionTime(transitionTime).
			WithSynchronizationState(apis.Failed)
		commands = []Command{*setStatus}
	} else {
		switch resource.GetStatus().Conditions.GetSyncStateFromReason() {
		case apis.Creating:
			commands = st.onCreating(
				ctx,
				resource,
				st.WorkflowRepository.GetByLabels(ctx, workflowconstants.CommonWorkflowLabels(resource)),
				transitionTime,
			)
		case apis.Succeeded, apis.Failed:
			if !resource.GetDeletionTimestamp().IsZero() {
				commands = st.onDelete(ctx, provider, providerSvc, resource, transitionTime)
			} else {
				commands = st.onSucceededOrFailed(ctx, provider, providerSvc, resource, transitionTime)
			}
		case apis.Updating:
			commands = st.onUpdating(
				ctx,
				resource,
				st.WorkflowRepository.GetByLabels(ctx, workflowconstants.CommonWorkflowLabels(resource)),
				transitionTime,
			)
		case apis.Deleting:
			commands = st.onDeleting(
				ctx,
				resource,
				st.WorkflowRepository.GetByLabels(ctx, workflowconstants.CommonWorkflowLabels(resource)),
				transitionTime,
			)
		case apis.Deleted:
		default:
			commands = st.onUnknown(ctx, provider, providerSvc, resource, transitionTime)
		}
	}

	if resource.GetStatus().Conditions.GetSyncStateFromReason() == apis.Deleted {
		commands = append([]Command{ReleaseResource{}}, commands...)
	} else {
		commands = append([]Command{AcquireResource{}}, commands...)
	}

	return
}

func (st *StateHandler[R]) StateTransition(
	ctx context.Context,
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
) []Command {
	logger := log.FromContext(ctx)
	logger.Info("state transition start")

	stateTransitionCommands := st.stateTransition(ctx, provider, providerSvc, resource, metav1.Now())
	return alwaysSetObservedGeneration(ctx, stateTransitionCommands, resource)
}

func (st *StateHandler[R]) onUnknown(
	ctx context.Context,
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
	transitionTime metav1.Time,
) []Command {
	logger := log.FromContext(ctx)

	newVersion := resource.ComputeVersion()

	if resource.GetStatus().Provider.Id != "" {
		logger.Info("empty state but ProviderId already exists, updating resource")
		workflow, err := st.WorkflowFactory.ConstructUpdateWorkflow(provider, providerSvc, resource)

		if err != nil {
			failureMessage := workflowconstants.ConstructionFailedError
			logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))

			return []Command{
				*From(resource.GetStatus()).
					WithVersion(newVersion).
					WithMessage(failureMessage).
					WithLastTransitionTime(transitionTime).
					WithSynchronizationState(apis.Failed),
			}
		}

		return []Command{
			*From(resource.GetStatus()).
				WithLastTransitionTime(transitionTime).
				WithSynchronizationState(apis.Updating).
				WithVersion(newVersion),
			CreateWorkflow{Workflow: *workflow},
		}
	}

	logger.Info("empty state, creating resource")

	workflow, err := st.WorkflowFactory.ConstructCreationWorkflow(provider, providerSvc, resource)

	if err != nil {
		failureMessage := workflowconstants.ConstructionFailedError
		logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))

		cmd := From(resource.GetStatus()).
			WithLastTransitionTime(transitionTime).
			WithMessage(failureMessage).
			WithSynchronizationState(apis.Failed).
			WithVersion(newVersion)

		return []Command{
			*cmd,
		}
	}

	status := SetStatus{
		Status: pipelineshub.Status{
			Version: newVersion,
		},
	}

	status.WithLastTransitionTime(transitionTime).WithSynchronizationState(apis.Creating)

	return []Command{
		status,
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st StateHandler[R]) onDelete(
	ctx context.Context,
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
	transitionTime metav1.Time,
) []Command {
	logger := log.FromContext(ctx)
	logger.Info("deletion requested, deleting")

	if resource.GetStatus().Provider.Id == "" {
		return []Command{
			*From(resource.GetStatus()).
				WithLastTransitionTime(transitionTime).
				WithSynchronizationState(apis.Deleted),
		}
	}

	workflow, err := st.WorkflowFactory.ConstructDeletionWorkflow(provider, providerSvc, resource)

	if err != nil {
		failureMessage := workflowconstants.ConstructionFailedError
		logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))

		return []Command{
			*From(resource.GetStatus()).
				WithMessage(failureMessage).
				WithLastTransitionTime(transitionTime).
				WithSynchronizationState(apis.Failed),
		}
	}

	return []Command{
		*From(resource.GetStatus()).
			WithLastTransitionTime(transitionTime).
			WithSynchronizationState(apis.Deleting),
		CreateWorkflow{Workflow: *workflow},
	}
}

func (st StateHandler[R]) onSucceededOrFailed(
	ctx context.Context,
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
	transitionTime metav1.Time,
) []Command {
	logger := log.FromContext(ctx)
	newResourceVersion := resource.ComputeVersion()

	if resource.GetStatus().Version == newResourceVersion {
		logger.V(2).Info("resource version has not changed")
		return []Command{}
	}

	var workflow *argo.Workflow
	var err error
	var targetState apis.SynchronizationState

	if resource.GetStatus().Provider.Id == "" {
		logger.Info("no providerId exists, creating")
		workflow, err = st.WorkflowFactory.ConstructCreationWorkflow(
			provider,
			providerSvc,
			resource,
		)

		if err != nil {
			failureMessage := workflowconstants.ConstructionFailedError
			logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))
			if wpErr, ok := err.(*workflowconstants.WorkflowParameterError); ok {
				failureMessage = wpErr.Error()
			}
			return []Command{
				*From(resource.GetStatus()).
					WithVersion(newResourceVersion).
					WithMessage(failureMessage).
					WithLastTransitionTime(transitionTime).
					WithSynchronizationState(apis.Failed),
			}
		}

		targetState = apis.Creating
	} else {
		logger.Info("providerId exists, updating")
		workflow, err = st.WorkflowFactory.ConstructUpdateWorkflow(provider, providerSvc, resource)

		if err != nil {
			failureMessage := workflowconstants.ConstructionFailedError
			logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))
			if wpErr, ok := err.(*workflowconstants.WorkflowParameterError); ok {
				failureMessage = wpErr.Error()
			}
			return []Command{
				*From(resource.GetStatus()).
					WithVersion(newResourceVersion).
					WithMessage(failureMessage).
					WithLastTransitionTime(transitionTime).
					WithSynchronizationState(apis.Failed),
			}
		}

		targetState = apis.Updating
	}

	return []Command{
		*From(resource.GetStatus()).
			WithLastTransitionTime(transitionTime).
			WithSynchronizationState(targetState).
			WithVersion(newResourceVersion),
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

func (st StateHandler[R]) setStateIfProviderFinished(
	ctx context.Context,
	status pipelineshub.Status,
	workflows []argo.Workflow,
	states IdVerifier,
	transitionTime metav1.Time,
) []Command {
	logger := log.FromContext(ctx)
	statusFromProviderOutput := func(workflow *argo.Workflow) *SetStatus {
		var handleWorkflowErr = func(err error) *SetStatus {
			failureMessage := "could not retrieve workflow output"
			logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))
			return From(status).
				WithMessage(failureMessage).
				WithLastTransitionTime(transitionTime).
				WithSynchronizationState(states.FailureState)
		}

		result, err := workflowutil.GetWorkflowOutput(workflow, workflowconstants.ProviderOutputParameterName)
		if err != nil {
			return handleWorkflowErr(err)
		}

		provider := workflowutil.GetWorkflowParameter(workflow, workflowconstants.ProviderNameParameterName)
		namespacedProvider, err := common.NamespacedNameFromString(provider)
		if err != nil {
			return handleWorkflowErr(err)
		}

		providerAndId := pipelineshub.ProviderAndId{Name: namespacedProvider, Id: result.Id}

		if result.ProviderError != "" {
			logger.Error(err, fmt.Sprintf("%s, failing resource", result.ProviderError))
			return From(status).
				WithMessage(result.ProviderError).
				WithProvider(providerAndId).
				WithLastTransitionTime(transitionTime).
				WithSynchronizationState(states.FailureState)
		}

		err = states.VerifyId(result.Id)

		if err != nil {
			failureMessage := err.Error()
			logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))
			return From(status).
				WithMessage(failureMessage).
				WithLastTransitionTime(transitionTime).
				WithSynchronizationState(states.FailureState)
		}

		return From(status).
			WithLastTransitionTime(transitionTime).
			WithSynchronizationState(states.SuccessState).
			WithProvider(providerAndId)
	}

	inProgress, succeeded, failed := workflowutil.LatestWorkflowByPhase(workflows)

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
		setStatusCommand = From(status).
			WithMessage(failureMessage).
			WithLastTransitionTime(transitionTime).
			WithSynchronizationState(states.FailureState)
	}

	return []Command{
		*setStatusCommand,
		MarkWorkflowsAsProcessed{
			Workflows: workflows,
		},
	}
}

func (st StateHandler[R]) onCreating(
	ctx context.Context,
	resource R,
	creationWorkflows []argo.Workflow,
	transitionTime metav1.Time,
) []Command {
	logger := log.FromContext(ctx)

	if resource.GetStatus().Version == "" {
		failureMessage := "creating resource with empty version"
		logger.Info(fmt.Sprintf("%s, failing resource", failureMessage))

		return []Command{
			*From(resource.GetStatus()).
				WithMessage(failureMessage).
				WithLastTransitionTime(transitionTime).
				WithSynchronizationState(apis.Failed),
		}
	}

	return st.setStateIfProviderFinished(ctx, resource.GetStatus(), creationWorkflows, succeedForEmptyId, transitionTime)
}

func (st StateHandler[R]) onUpdating(
	ctx context.Context,
	resource R,
	updateWorkflows []argo.Workflow,
	transitionTime metav1.Time,
) []Command {
	logger := log.FromContext(ctx)

	if resource.GetStatus().Version == "" || resource.GetStatus().Provider.Id == "" {
		failureMessage := "updating resource with empty version or providerId"
		logger.Info(fmt.Sprintf("%s, failing resource", failureMessage))

		return []Command{
			*From(resource.GetStatus()).
				WithMessage(failureMessage).
				WithLastTransitionTime(transitionTime).
				WithSynchronizationState(apis.Failed),
		}
	}

	return st.setStateIfProviderFinished(ctx, resource.GetStatus(), updateWorkflows, succeedForEmptyId, transitionTime)
}

func (st StateHandler[R]) onDeleting(
	ctx context.Context,
	resource R,
	deletionWorkflows []argo.Workflow,
	transitionTime metav1.Time,
) []Command {
	return st.setStateIfProviderFinished(ctx, resource.GetStatus(), deletionWorkflows, deletedForNonEmptyId, transitionTime)
}
