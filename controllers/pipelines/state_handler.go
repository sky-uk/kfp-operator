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
			WithSyncStateCondition(apis.Failed, transitionTime, StateHandlerConstants.ProviderChangedError)
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
	logger.V(2).Info("state transition start")

	time := metav1.Now()
	stateTransitionCommands := st.stateTransition(ctx, provider, providerSvc, resource, time)
	return alwaysSetObservedGeneration(ctx, stateTransitionCommands, resource, time)
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
					WithSyncStateCondition(apis.Failed, transitionTime, failureMessage),
			}
		}

		return []Command{
			*From(resource.GetStatus()).
				WithSyncStateCondition(apis.Updating, transitionTime, "").
				WithVersion(newVersion),
			CreateWorkflow{Workflow: *workflow},
		}
	}

	logger.V(2).Info("empty state, creating resource")

	workflow, err := st.WorkflowFactory.ConstructCreationWorkflow(provider, providerSvc, resource)

	if err != nil {
		failureMessage := workflowconstants.ConstructionFailedError
		logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))

		status := From(resource.GetStatus()).
			WithSyncStateCondition(apis.Failed, transitionTime, failureMessage).
			WithVersion(newVersion)

		return []Command{
			*status,
		}
	}

	status := SetStatus{
		Status: pipelineshub.Status{
			Version: newVersion,
		},
	}

	status.WithSyncStateCondition(apis.Creating, transitionTime, "")

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
	logger.V(2).Info("deletion requested, deleting")

	if resource.GetStatus().Provider.Id == "" {
		return []Command{
			*From(resource.GetStatus()).
				WithSyncStateCondition(apis.Deleted, transitionTime, ""),
		}
	}

	workflow, err := st.WorkflowFactory.ConstructDeletionWorkflow(provider, providerSvc, resource)

	if err != nil {
		failureMessage := workflowconstants.ConstructionFailedError
		logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))

		return []Command{
			*From(resource.GetStatus()).
				WithSyncStateCondition(apis.Failed, transitionTime, failureMessage),
		}
	}

	return []Command{
		*From(resource.GetStatus()).
			WithSyncStateCondition(apis.Deleting, transitionTime, ""),
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

	if resource.GetStatus().Version == newResourceVersion && resource.GetStatus().Conditions.GetSyncStateFromReason() != apis.Failed {
		logger.V(2).Info("resource version has not changed")
		return []Command{}
	}

	var workflow *argo.Workflow
	var err error
	var targetState apis.SynchronizationState

	providerId := resource.GetStatus().Provider.Id
	if providerId == "" {
		logger.V(2).Info("no providerId exists, creating")
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
					WithSyncStateCondition(apis.Failed, transitionTime, failureMessage),
			}
		}

		targetState = apis.Creating
	} else {
		logger.V(2).Info("providerId exists, updating", "providerId", providerId)
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
					WithSyncStateCondition(apis.Failed, transitionTime, failureMessage),
			}
		}

		targetState = apis.Updating
	}

	return []Command{
		*From(resource.GetStatus()).
			WithSyncStateCondition(targetState, transitionTime, "").
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
				WithSyncStateCondition(states.FailureState, transitionTime, failureMessage)
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
				WithProvider(providerAndId).
				WithSyncStateCondition(states.FailureState, transitionTime, result.ProviderError)
		}

		err = states.VerifyId(result.Id)

		if err != nil {
			failureMessage := err.Error()
			logger.Error(err, fmt.Sprintf("%s, failing resource", failureMessage))
			return From(status).
				WithSyncStateCondition(states.FailureState, transitionTime, failureMessage)
		}

		return From(status).
			WithSyncStateCondition(states.SuccessState, transitionTime, "").
			WithProvider(providerAndId)
	}

	inProgress, succeeded, failed := workflowutil.LatestWorkflowByPhase(workflows)

	if inProgress != nil {
		logger.V(2).Info("operation in progress")
		return []Command{}
	}

	var setStatusCommand *SetStatus

	if succeeded != nil {
		logger.V(2).Info("operation succeeded")
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
			WithSyncStateCondition(states.FailureState, transitionTime, failureMessage)
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
				WithSyncStateCondition(apis.Failed, transitionTime, failureMessage),
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
				WithSyncStateCondition(apis.Failed, transitionTime, failureMessage),
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
