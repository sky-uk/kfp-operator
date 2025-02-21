//go:build unit

package pipelines

import (
	"context"
	"fmt"
	"time"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StateTransitionTestCase struct {
	workflowFactory *TestWorkflowFactory
	Experiment      *pipelinesv1.TestResource
	Service         *corev1.Service
	SystemStatus    StubbedWorkflows
	Commands        []Command
}

type TestWorkflowFactory struct {
	CalledWithProvider *pipelinesv1.Provider
	shouldFail         bool
}

func (f *TestWorkflowFactory) ConstructCreationWorkflow(
	provider pipelinesv1.Provider,
	_ corev1.Service,
	_ *pipelinesv1.TestResource,
) (*argo.Workflow, error) {
	f.CalledWithProvider = &provider
	if f.shouldFail {
		return nil, fmt.Errorf("an error occurred")
	}
	return &argo.Workflow{}, nil
}

func (f *TestWorkflowFactory) ConstructUpdateWorkflow(
	provider pipelinesv1.Provider,
	_ corev1.Service,
	_ *pipelinesv1.TestResource,
) (*argo.Workflow, error) {
	f.CalledWithProvider = &provider
	if f.shouldFail {
		return nil, fmt.Errorf("an error occurred")
	}
	return &argo.Workflow{}, nil
}

func (f *TestWorkflowFactory) ConstructDeletionWorkflow(
	provider pipelinesv1.Provider,
	_ corev1.Service,
	_ *pipelinesv1.TestResource,
) (*argo.Workflow, error) {
	f.CalledWithProvider = &provider
	if f.shouldFail {
		return nil, fmt.Errorf("an error occurred")
	}
	return &argo.Workflow{}, nil
}

func (st StateTransitionTestCase) WorkflowConstructionFails() StateTransitionTestCase {
	st.workflowFactory.shouldFail = true
	return st
}

func (st StateTransitionTestCase) WithWorkFlow(workflow *argo.Workflow) StateTransitionTestCase {
	st.SystemStatus.AddWorkflow(*workflow)
	return st
}

func (st StateTransitionTestCase) WithCreateWorkFlow(phase argo.WorkflowPhase) StateTransitionTestCase {
	return st.WithWorkFlow(CreateTestWorkflow(phase))
}

func (st StateTransitionTestCase) WithSucceededCreateWorkFlow(
	provider pipelinesv1.Provider,
	providerId pipelinesv1.ProviderAndId,
	providerError string,
) StateTransitionTestCase {
	workflow, err := workflowutil.SetWorkflowProvider(
		CreateTestWorkflow(argo.WorkflowSucceeded),
		provider)
	Expect(err).NotTo(HaveOccurred())
	return st.WithWorkFlow(
		workflowutil.SetProviderOutput(
			workflow,
			providers.Output{Id: providerId.Id, ProviderError: providerError},
		),
	)
}

func (st StateTransitionTestCase) WithFailedUpdateWorkflow() StateTransitionTestCase {
	return st.WithWorkFlow(
		CreateTestWorkflow(argo.WorkflowFailed),
	)
}

func (st StateTransitionTestCase) WithSucceededUpdateWorkflow(
	provider pipelinesv1.Provider,
	providerId pipelinesv1.ProviderAndId,
	providerError string,
) StateTransitionTestCase {
	workflow, err := workflowutil.SetWorkflowProvider(
		CreateTestWorkflow(argo.WorkflowSucceeded),
		provider)
	Expect(err).NotTo(HaveOccurred())
	return st.WithWorkFlow(
		workflowutil.SetProviderOutput(
			workflow,
			providers.Output{Id: providerId.Id, ProviderError: providerError},
		),
	)
}

func (st StateTransitionTestCase) WithDeletionWorkflow(phase argo.WorkflowPhase) StateTransitionTestCase {
	return st.WithWorkFlow(
		CreateTestWorkflow(phase),
	)
}

func (st StateTransitionTestCase) WithSucceededDeletionWorkflow(
	provider pipelinesv1.Provider,
	providerId pipelinesv1.ProviderAndId,
	providerError string,
) StateTransitionTestCase {
	workflow, err := workflowutil.SetWorkflowProvider(
		CreateTestWorkflow(argo.WorkflowSucceeded),
		provider)
	Expect(err).NotTo(HaveOccurred())
	return st.WithWorkFlow(
		workflowutil.SetProviderOutput(
			workflow,
			providers.Output{Id: providerId.Id, ProviderError: providerError},
		),
	)
}

func (st StateTransitionTestCase) IssuesCreationWorkflow() StateTransitionTestCase {
	creationWorkflow, _ := st.workflowFactory.ConstructCreationWorkflow(
		*pipelinesv1.RandomProvider(),
		*st.Service,
		st.Experiment,
	)
	return st.IssuesCommand(CreateWorkflow{Workflow: *creationWorkflow})
}

func (st StateTransitionTestCase) IssuesUpdateWorkflow() StateTransitionTestCase {
	updateWorkflow, _ := st.workflowFactory.ConstructUpdateWorkflow(
		*pipelinesv1.RandomProvider(),
		*st.Service,
		st.Experiment,
	)
	return st.IssuesCommand(CreateWorkflow{Workflow: *updateWorkflow})
}

func (st StateTransitionTestCase) IssuesDeletionWorkflow() StateTransitionTestCase {
	deletionWorkflow, _ := st.workflowFactory.ConstructDeletionWorkflow(
		*pipelinesv1.RandomProvider(),
		*st.Service,
		st.Experiment,
	)
	return st.IssuesCommand(CreateWorkflow{Workflow: *deletionWorkflow})
}

func (st StateTransitionTestCase) MarksAllWorkflowsAsProcessed() StateTransitionTestCase {
	return st.IssuesCommand(MarkWorkflowsAsProcessed{
		Workflows: st.SystemStatus.Workflows,
	})
}

func (st StateTransitionTestCase) AcquireExperiment() StateTransitionTestCase {
	return st.IssuesCommand(AcquireResource{})
}

func (st StateTransitionTestCase) ReleaseExperiment() StateTransitionTestCase {
	return st.IssuesCommand(ReleaseResource{})
}

func (st StateTransitionTestCase) IssuesCommand(command Command) StateTransitionTestCase {
	st.Commands = append(st.Commands, command)
	return st
}

func (st StateTransitionTestCase) DeletionRequested() StateTransitionTestCase {
	st.Experiment.DeletionTimestamp = &metav1.Time{
		Time: time.UnixMilli(1),
	}
	return st
}

func anyNonDeletedState() apis.SynchronizationState {
	for {
		if state := apis.RandomSynchronizationState(); state != apis.Deleted {
			return state
		}
	}
}

var _ = Describe("State handler", func() {
	provider := pipelinesv1.RandomProvider()
	providerId := pipelinesv1.ProviderAndId{
		Name: provider.Name,
		Id:   apis.RandomString(),
	}
	anotherIdSameProvider := pipelinesv1.ProviderAndId{
		Name: provider.Name,
		Id:   apis.RandomString(),
	}
	anotherProviderId := pipelinesv1.ProviderAndId{
		Name: apis.RandomString(),
		Id:   apis.RandomString(),
	}
	emptyProviderId := pipelinesv1.ProviderAndId{}
	emptyProvider := pipelinesv1.RandomProvider()
	emptyProvider.Name = ""

	providerError := "a provider error has occurred"
	irrelevant := "irrelevant"
	v1 := apis.RandomShortHash()
	v2 := apis.RandomShortHash()
	UnknownState := apis.SynchronizationState(apis.RandomString())

	var Check = func(description string, transition StateTransitionTestCase) TableEntry {
		return Entry(
			description,
			transition,
		)
	}

	var From = func(
		status apis.SynchronizationState,
		id pipelinesv1.ProviderAndId,
		versionInState string,
		computedVersion string,
	) StateTransitionTestCase {
		resource := pipelinesv1.RandomResource()
		resource.SetStatus(pipelinesv1.Status{
			SynchronizationState: status,
			Version:              versionInState,
			Provider:             id,
		})
		resource.SetComputedVersion(computedVersion)

		return StateTransitionTestCase{
			workflowFactory: &TestWorkflowFactory{}, // TODO: mock workflowFactory
			Experiment:      resource,
			Service:         &corev1.Service{},
			Commands:        []Command{},
		}
	}

	DescribeTable("State transitions", func(st StateTransitionTestCase) {
		var stateHandler = StateHandler[*pipelinesv1.TestResource]{
			WorkflowRepository: st.SystemStatus,
			WorkflowFactory:    st.workflowFactory,
		}
		commands := stateHandler.stateTransition(context.Background(), *provider, *st.Service, st.Experiment)
		Expect(commands).To(Equal(st.Commands))
		if st.workflowFactory.CalledWithProvider != nil {
			Expect(st.workflowFactory.CalledWithProvider).To(BeComparableTo(provider))
		}
	},
		Check("Empty",
			From(UnknownState, emptyProviderId, "", v1).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Empty and workflow creation fails",
			From(UnknownState, emptyProviderId, "", v1).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v1).
					WithMessage(workflowconstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Empty with version",
			From(UnknownState, emptyProviderId, v1, v1).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Empty with id",
			From(UnknownState, providerId, "", v1).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithProvider(providerId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Empty with id and workflow creation fails",
			From(UnknownState, providerId, "", v1).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithProvider(providerId).
					WithVersion(v1).
					WithMessage(workflowconstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Empty with id and version",
			From(UnknownState, providerId, v1, v2).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithProvider(providerId).
					WithVersion(v2)).
				IssuesUpdateWorkflow(),
		),
		Check("Creating succeeds",
			From(apis.Creating, emptyProviderId, v1, irrelevant).
				AcquireExperiment().
				WithSucceededCreateWorkFlow(*provider, providerId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Succeeded).
					WithProvider(providerId).
					WithVersion(v1)).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Creating succeeds without providerId or provider error",
			From(apis.Creating, providerId, v1, irrelevant).
				AcquireExperiment().
				WithSucceededCreateWorkFlow(*provider, emptyProviderId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProvider(providerId).
					WithVersion(v1).
					WithMessage("id was empty")).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Creating succeeds with provider error",
			From(apis.Creating, anotherIdSameProvider, v1, irrelevant).
				AcquireExperiment().
				WithSucceededCreateWorkFlow(*provider, providerId, providerError).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProvider(providerId).
					WithVersion(v1).
					WithMessage(providerError)).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Creating fails",
			From(apis.Creating, emptyProviderId, v1, irrelevant).
				AcquireExperiment().
				WithCreateWorkFlow(argo.WorkflowFailed).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithVersion(v1).
					WithMessage("operation failed")).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Creating without version",
			From(apis.Creating, emptyProviderId, "", irrelevant).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithMessage("creating resource with empty version")),
		),
		Check("Succeeded no update",
			From(apis.Succeeded, providerId, v1, v1).
				AcquireExperiment(),
		),
		Check("Succeeded with update",
			From(apis.Succeeded, providerId, v1, v2).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithProvider(providerId).
					WithVersion(v2)).
				IssuesUpdateWorkflow(),
		),
		Check("Succeeded with update and workflow creation fails",
			From(apis.Succeeded, providerId, v1, v2).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithProvider(providerId).
					WithVersion(v2).
					WithMessage(workflowconstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Succeeded with update but no ProviderAndId",
			From(apis.Succeeded, emptyProviderId, v1, v2).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v2)).
				IssuesCreationWorkflow(),
		),
		Check("Succeeded with update and workflow creation fails",
			From(apis.Succeeded, emptyProviderId, v1, v2).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v2).
					WithMessage(workflowconstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Succeeded with update but no ProviderAndId and no version",
			From(apis.Succeeded, emptyProviderId, "", v1).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Failed no update",
			From(apis.Failed, providerId, v1, v1).
				AcquireExperiment(),
		),
		Check("Failed with update",
			From(apis.Failed, providerId, v1, v2).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithProvider(providerId).
					WithVersion(v2)).
				IssuesUpdateWorkflow(),
		),
		Check("Failed with update and workflow creation fails",
			From(apis.Failed, providerId, v1, v2).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithProvider(providerId).
					WithVersion(v2).
					WithMessage(workflowconstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Failed with update but no ProviderAndId",
			From(apis.Failed, emptyProviderId, v1, v2).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v2)).
				IssuesCreationWorkflow(),
		),
		Check("Failed with update but no ProviderAndId and workflow creation fails",
			From(apis.Failed, emptyProviderId, v1, v2).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v2).
					WithMessage(workflowconstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Failed with update but no ProviderAndId and no version",
			From(apis.Failed, emptyProviderId, "", v1).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Updating succeeds with providerId",
			From(apis.Updating, providerId, v1, irrelevant).
				AcquireExperiment().
				WithSucceededUpdateWorkflow(*provider, anotherIdSameProvider, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Succeeded).
					WithProvider(anotherIdSameProvider).
					WithVersion(v1)).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Updating succeeds without providerId or provider error",
			From(apis.Updating, providerId, v1, irrelevant).
				AcquireExperiment().
				WithSucceededUpdateWorkflow(*provider, emptyProviderId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProvider(providerId).
					WithVersion(v1).
					WithMessage("id was empty")).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Updating succeeds with provider error",
			From(apis.Updating, providerId, v1, irrelevant).
				AcquireExperiment().
				WithSucceededUpdateWorkflow(*provider, anotherIdSameProvider, providerError).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProvider(anotherIdSameProvider).
					WithVersion(v1).
					WithMessage(providerError)).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Updating fails",
			From(apis.Updating, providerId, v1, irrelevant).
				AcquireExperiment().
				WithFailedUpdateWorkflow().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProvider(providerId).
					WithVersion(v1).
					WithMessage("operation failed")).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Updating without version",
			From(apis.Updating, providerId, "", irrelevant).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProvider(providerId).
					WithMessage("updating resource with empty version or providerId")),
		),
		Check("Updating without ProviderAndId",
			From(apis.Updating, emptyProviderId, v1, irrelevant).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithVersion(v1).
					WithMessage("updating resource with empty version or providerId")),
		),
		Check("Updating without ProviderAndId or version",
			From(apis.Updating, emptyProviderId, "", irrelevant).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithMessage("updating resource with empty version or providerId")),
		),
		Check("Deleting from Succeeded",
			From(apis.Succeeded, providerId, v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithProvider(providerId).
					WithVersion(v1)).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Succeeded without providerId",
			From(apis.Succeeded, emptyProviderId, v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleted).
					WithVersion(v1)),
		),
		Check("Deleting from Failed",
			From(apis.Failed, providerId, v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithProvider(providerId).
					WithVersion(v1)).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Failed without providerId",
			From(apis.Failed, emptyProviderId, v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleted).
					WithVersion(v1)),
		),
		Check("Deletion succeeds with providerId",
			From(apis.Deleting, providerId, v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				WithSucceededDeletionWorkflow(*provider, anotherIdSameProvider, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithProvider(providerId).
					WithVersion(v1).
					WithMessage("id should be empty")).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Deletion succeeds",
			From(apis.Deleting, providerId, v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				WithSucceededDeletionWorkflow(*emptyProvider, emptyProviderId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleted).
					WithProvider(emptyProviderId).
					WithVersion(v1)).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Deletion succeeds with provider error",
			From(apis.Deleting, providerId, v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				WithSucceededDeletionWorkflow(*provider, providerId, providerError).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithProvider(providerId).
					WithVersion(v1).
					WithMessage(providerError)).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Deletion fails",
			From(apis.Deleting, providerId, v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				WithDeletionWorkflow(argo.WorkflowFailed).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithProvider(providerId).
					WithVersion(v1).
					WithMessage("operation failed")).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Stay in deleted",
			From(apis.Deleted, providerId, v1, irrelevant).
				ReleaseExperiment(),
		),
		Check("Any non-deleted state with different provider",
			From(anyNonDeletedState(), anotherProviderId, irrelevant, irrelevant).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithVersion(irrelevant).
					WithProvider(anotherProviderId).
					WithMessage(StateHandlerConstants.ProviderChangedError).
					WithSynchronizationState(apis.Failed)),
		))
})
