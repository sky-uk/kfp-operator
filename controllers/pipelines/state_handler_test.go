//go:build unit
// +build unit

package pipelines

import (
	"context"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	providers "github.com/sky-uk/kfp-operator/providers/base"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type StateTransitionTestCase struct {
	workflowFactory WorkflowFactory[*pipelinesv1.TestResource]
	Experiment      *pipelinesv1.TestResource
	SystemStatus    StubbedWorkflows
	Commands        []Command
}

type SucceedingWorkflowFactory struct {
	CalledWithProvider string
}

func (f *SucceedingWorkflowFactory) ConstructCreationWorkflow(provider string, _ *pipelinesv1.TestResource) (*argo.Workflow, error) {
	f.CalledWithProvider = provider
	return &argo.Workflow{}, nil
}

func (f *SucceedingWorkflowFactory) ConstructUpdateWorkflow(provider string, _ *pipelinesv1.TestResource) (*argo.Workflow, error) {
	f.CalledWithProvider = provider
	return &argo.Workflow{}, nil
}

func (f *SucceedingWorkflowFactory) ConstructDeletionWorkflow(provider string, _ *pipelinesv1.TestResource) (*argo.Workflow, error) {
	f.CalledWithProvider = provider
	return &argo.Workflow{}, nil
}

type FailingWorkflowFactory struct{}

func (f FailingWorkflowFactory) ConstructCreationWorkflow(_ string, _ *pipelinesv1.TestResource) (*argo.Workflow, error) {
	return nil, fmt.Errorf("an error occurred")
}

func (f FailingWorkflowFactory) ConstructUpdateWorkflow(_ string, _ *pipelinesv1.TestResource) (*argo.Workflow, error) {
	return nil, fmt.Errorf("an error occurred")
}

func (f FailingWorkflowFactory) ConstructDeletionWorkflow(_ string, _ *pipelinesv1.TestResource) (*argo.Workflow, error) {
	return nil, fmt.Errorf("an error occurred")
}

func (st StateTransitionTestCase) WorkflowConstructionFails() StateTransitionTestCase {
	st.workflowFactory = FailingWorkflowFactory{}
	return st
}

func (st StateTransitionTestCase) WithWorkFlow(workflow *argo.Workflow) StateTransitionTestCase {
	st.SystemStatus.AddWorkflow(*workflow)
	return st
}

func (st StateTransitionTestCase) WithCreateWorkFlow(phase argo.WorkflowPhase) StateTransitionTestCase {
	return st.WithWorkFlow(CreateTestWorkflow(WorkflowConstants.CreateOperationLabel, phase))
}

func (st StateTransitionTestCase) WithSucceededCreateWorkFlow(providerId pipelinesv1.ProviderAndId, providerError string) StateTransitionTestCase {
	return st.WithWorkFlow(
		setProviderOutput(
			setWorkflowProvider(
				CreateTestWorkflow(WorkflowConstants.CreateOperationLabel, argo.WorkflowSucceeded),
				providerId.Provider),
			providers.Output{Id: providerId.Id, ProviderError: providerError},
		),
	)
}

func (st StateTransitionTestCase) WithFailedUpdateWorkflow() StateTransitionTestCase {
	return st.WithWorkFlow(
		CreateTestWorkflow(WorkflowConstants.UpdateOperationLabel, argo.WorkflowFailed),
	)
}

func (st StateTransitionTestCase) WithSucceededUpdateWorkflow(providerId pipelinesv1.ProviderAndId, providerError string) StateTransitionTestCase {
	return st.WithWorkFlow(
		setProviderOutput(
			setWorkflowProvider(
				CreateTestWorkflow(WorkflowConstants.UpdateOperationLabel, argo.WorkflowSucceeded),
				providerId.Provider),
			providers.Output{Id: providerId.Id, ProviderError: providerError},
		),
	)
}

func (st StateTransitionTestCase) WithDeletionWorkflow(phase argo.WorkflowPhase) StateTransitionTestCase {
	return st.WithWorkFlow(
		CreateTestWorkflow(WorkflowConstants.DeleteOperationLabel, phase),
	)
}

func (st StateTransitionTestCase) WithSucceededDeletionWorkflow(providerId pipelinesv1.ProviderAndId, providerError string) StateTransitionTestCase {
	return st.WithWorkFlow(
		setProviderOutput(
			setWorkflowProvider(
				CreateTestWorkflow(WorkflowConstants.DeleteOperationLabel, argo.WorkflowSucceeded),
				providerId.Provider),
			providers.Output{Id: providerId.Id, ProviderError: providerError},
		),
	)
}

func (st StateTransitionTestCase) IssuesCreationWorkflow() StateTransitionTestCase {
	creationWorkflow, _ := st.workflowFactory.ConstructCreationWorkflow(apis.RandomString(), st.Experiment)
	return st.IssuesCommand(CreateWorkflow{Workflow: *creationWorkflow})
}

func (st StateTransitionTestCase) IssuesUpdateWorkflow() StateTransitionTestCase {
	updateWorkflow, _ := st.workflowFactory.ConstructUpdateWorkflow(apis.RandomString(), st.Experiment)
	return st.IssuesCommand(CreateWorkflow{Workflow: *updateWorkflow})
}

func (st StateTransitionTestCase) IssuesDeletionWorkflow() StateTransitionTestCase {
	deletionWorkflow, _ := st.workflowFactory.ConstructDeletionWorkflow(apis.RandomString(), st.Experiment)
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
	st.Experiment.DeletionTimestamp = &metav1.Time{time.UnixMilli(1)}
	return st
}

var _ = Describe("State handler", func() {
	provider := apis.RandomString()
	providerId := pipelinesv1.ProviderAndId{
		Provider: provider,
		Id:       apis.RandomString(),
	}
	anotherProviderId := pipelinesv1.ProviderAndId{
		Provider: provider,
		Id:       apis.RandomString(),
	}
	emptyProviderId := pipelinesv1.ProviderAndId{}

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

	var From = func(status apis.SynchronizationState, id pipelinesv1.ProviderAndId, versionInState string, computedVersion string) StateTransitionTestCase {
		resource := pipelinesv1.RandomResource()
		resource.SetStatus(pipelinesv1.Status{
			SynchronizationState: status,
			Version:              versionInState,
			ProviderId:           id,
		})
		resource.SetComputedVersion(computedVersion)

		return StateTransitionTestCase{
			workflowFactory: &SucceedingWorkflowFactory{}, // TODO: mock workflowFactory
			Experiment:      resource,
			Commands:        []Command{},
		}
	}

	DescribeTable("State transitions", func(st StateTransitionTestCase) {
		var stateHandler = StateHandler[*pipelinesv1.TestResource]{
			WorkflowRepository: st.SystemStatus,
			WorkflowFactory:    st.workflowFactory,
		}
		commands := stateHandler.stateTransition(context.Background(), st.Experiment)
		Expect(commands).To(Equal(st.Commands))
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
					WithMessage(WorkflowConstants.ConstructionFailedError).
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
					WithProviderId(providerId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Empty with id and workflow creation fails",
			From(UnknownState, providerId, "", v1).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithProviderId(providerId).
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Empty with id and version",
			From(UnknownState, providerId, v1, v2).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithProviderId(providerId).
					WithVersion(v2)).
				IssuesUpdateWorkflow(),
		),
		Check("Creating succeeds",
			From(apis.Creating, emptyProviderId, v1, irrelevant).
				AcquireExperiment().
				WithSucceededCreateWorkFlow(providerId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Succeeded).
					WithProviderId(providerId).
					WithVersion(v1)).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Creating succeeds without providerId or provider error",
			From(apis.Creating, providerId, v1, irrelevant).
				AcquireExperiment().
				WithSucceededCreateWorkFlow(emptyProviderId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProviderId(providerId).
					WithVersion(v1).
					WithMessage("id was empty")).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Creating succeeds with provider error",
			From(apis.Creating, anotherProviderId, v1, irrelevant).
				AcquireExperiment().
				WithSucceededCreateWorkFlow(providerId, providerError).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProviderId(providerId).
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
					WithProviderId(providerId).
					WithVersion(v2)).
				IssuesUpdateWorkflow(),
		),
		Check("Succeeded with update and workflow creation fails",
			From(apis.Succeeded, providerId, v1, v2).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithProviderId(providerId).
					WithVersion(v2).
					WithMessage(WorkflowConstants.ConstructionFailedError).
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
					WithMessage(WorkflowConstants.ConstructionFailedError).
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
					WithProviderId(providerId).
					WithVersion(v2)).
				IssuesUpdateWorkflow(),
		),
		Check("Failed with update and workflow creation fails",
			From(apis.Failed, providerId, v1, v2).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithProviderId(providerId).
					WithVersion(v2).
					WithMessage(WorkflowConstants.ConstructionFailedError).
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
					WithMessage(WorkflowConstants.ConstructionFailedError).
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
				WithSucceededUpdateWorkflow(anotherProviderId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Succeeded).
					WithProviderId(anotherProviderId).
					WithVersion(v1)).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Updating succeeds without providerId or provider error",
			From(apis.Updating, providerId, v1, irrelevant).
				AcquireExperiment().
				WithSucceededUpdateWorkflow(emptyProviderId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProviderId(providerId).
					WithVersion(v1).
					WithMessage("id was empty")).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Updating succeeds with provider error",
			From(apis.Updating, providerId, v1, irrelevant).
				AcquireExperiment().
				WithSucceededUpdateWorkflow(anotherProviderId, providerError).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProviderId(anotherProviderId).
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
					WithProviderId(providerId).
					WithVersion(v1).
					WithMessage("operation failed")).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Updating without version",
			From(apis.Updating, providerId, "", irrelevant).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProviderId(providerId).
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
					WithProviderId(providerId).
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
					WithProviderId(providerId).
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
				WithSucceededDeletionWorkflow(anotherProviderId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithProviderId(providerId).
					WithVersion(v1).
					WithMessage("id should be empty")).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Deletion succeeds",
			From(apis.Deleting, providerId, v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				WithSucceededDeletionWorkflow(emptyProviderId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleted).
					WithProviderId(emptyProviderId).
					WithVersion(v1)).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Deletion succeeds with provider error",
			From(apis.Deleting, providerId, v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				WithSucceededDeletionWorkflow(providerId, providerError).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithProviderId(providerId).
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
					WithProviderId(providerId).
					WithVersion(v1).
					WithMessage("operation failed")).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Stay in deleted",
			From(apis.Deleted, providerId, v1, irrelevant).
				ReleaseExperiment(),
		))

	_ = It("Creates with default provider", func() {
		provider := apis.RandomString()
		mockFactory := SucceedingWorkflowFactory{}
		var stateHandler = StateHandler[*pipelinesv1.TestResource]{
			WorkflowFactory: &mockFactory,
			Config: config.Configuration{
				DefaultProvider: provider,
			},
		}

		testResource := &pipelinesv1.TestResource{}
		stateHandler.stateTransition(context.Background(), testResource)
		Expect(mockFactory.CalledWithProvider).To(Equal(provider))
	})

	_ = It("Creates with provider from annotation", func() {
		provider := apis.RandomString()
		mockFactory := SucceedingWorkflowFactory{}
		var stateHandler = StateHandler[*pipelinesv1.TestResource]{
			WorkflowFactory: &mockFactory,
		}

		testResource := &pipelinesv1.TestResource{}
		testResource.SetAnnotations(map[string]string{
			"this should be filled in on Monday ;)": provider,
		})
		stateHandler.stateTransition(context.Background(), testResource)
		Expect(mockFactory.CalledWithProvider).To(Equal(provider))
	})
})
