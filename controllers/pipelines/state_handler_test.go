//go:build unit

package pipelines

import (
	"context"
	"fmt"

	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	"time"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StateTransitionTestCase struct {
	workflowFactory    *TestWorkflowFactory
	Experiment         *pipelineshub.TestResource
	SystemStatus       StubbedWorkflows
	LastTransitionTime metav1.Time
	Commands           []Command
}

type TestWorkflowFactory struct {
	CalledWithProvider    *pipelineshub.Provider
	CalledWithProviderSvc *corev1.Service
	shouldFail            bool
}

func (f *TestWorkflowFactory) ConstructCreationWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	_ *pipelineshub.TestResource,
) (*argo.Workflow, error) {
	f.CalledWithProvider = &provider
	f.CalledWithProviderSvc = &providerSvc
	if f.shouldFail {
		return nil, fmt.Errorf("an error occurred")
	}
	return &argo.Workflow{}, nil
}

func (f *TestWorkflowFactory) ConstructUpdateWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	_ *pipelineshub.TestResource,
) (*argo.Workflow, error) {
	f.CalledWithProvider = &provider
	f.CalledWithProviderSvc = &providerSvc
	if f.shouldFail {
		return nil, fmt.Errorf("an error occurred")
	}
	return &argo.Workflow{}, nil
}

func (f *TestWorkflowFactory) ConstructDeletionWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	_ *pipelineshub.TestResource,
) (*argo.Workflow, error) {
	f.CalledWithProvider = &provider
	f.CalledWithProviderSvc = &providerSvc
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
	provider pipelineshub.Provider,
	providerId pipelineshub.ProviderAndId,
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
	provider pipelineshub.Provider,
	providerId pipelineshub.ProviderAndId,
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
	provider pipelineshub.Provider,
	providerId pipelineshub.ProviderAndId,
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
		*pipelineshub.RandomProvider(),
		*RandomProviderService(),
		st.Experiment,
	)
	return st.IssuesCommand(CreateWorkflow{Workflow: *creationWorkflow})
}

func (st StateTransitionTestCase) IssuesUpdateWorkflow() StateTransitionTestCase {
	updateWorkflow, _ := st.workflowFactory.ConstructUpdateWorkflow(
		*pipelineshub.RandomProvider(),
		*RandomProviderService(),
		st.Experiment,
	)
	return st.IssuesCommand(CreateWorkflow{Workflow: *updateWorkflow})
}

func (st StateTransitionTestCase) IssuesDeletionWorkflow() StateTransitionTestCase {
	deletionWorkflow, _ := st.workflowFactory.ConstructDeletionWorkflow(
		*pipelineshub.RandomProvider(),
		*RandomProviderService(),
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

func RandomProviderService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apis.RandomLowercaseString(),
			Namespace: apis.RandomLowercaseString(),
		},
	}
}

var _ = Describe("State handler", func() {
	provider := pipelineshub.RandomProvider()
	providerSvc := RandomProviderService()
	providerId := pipelineshub.ProviderAndId{
		Name: common.NamespacedName{
			Name:      provider.Name,
			Namespace: provider.Namespace,
		},
		Id: apis.RandomString(),
	}
	anotherIdSameProvider := pipelineshub.ProviderAndId{
		Name: common.NamespacedName{
			Name:      provider.Name,
			Namespace: provider.Namespace,
		},
		Id: apis.RandomString(),
	}
	anotherProviderId := pipelineshub.ProviderAndId{
		Name: common.NamespacedName{
			Name:      apis.RandomString(),
			Namespace: apis.RandomString(),
		},
		Id: apis.RandomString(),
	}
	emptyProviderId := pipelineshub.ProviderAndId{}
	emptyProvider := pipelineshub.RandomProvider()
	emptyProvider.Name = ""
	emptyProvider.Namespace = ""

	transitionTime := metav1.Now()

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
		id pipelineshub.ProviderAndId,
		versionInState string,
		computedVersion string,
		transitionTime metav1.Time,
	) StateTransitionTestCase {
		resource := pipelineshub.RandomResource()
		resource.SetStatus(pipelineshub.Status{
			Version:  versionInState,
			Provider: id,
			Conditions: apis.Conditions{
				{
					Type:               apis.ConditionTypes.SynchronizationSucceeded,
					Status:             apis.ConditionStatusForSynchronizationState(status),
					Reason:             string(status),
					LastTransitionTime: transitionTime,
				},
			},
		})
		resource.SetComputedVersion(computedVersion)

		return StateTransitionTestCase{
			workflowFactory:    &TestWorkflowFactory{}, // TODO: mock workflowFactory
			Experiment:         resource,
			LastTransitionTime: transitionTime,
			Commands:           []Command{},
		}
	}

	DescribeTable("State transitions", func(st StateTransitionTestCase) {
		var stateHandler = StateHandler[*pipelineshub.TestResource]{
			WorkflowRepository: st.SystemStatus,
			WorkflowFactory:    st.workflowFactory,
		}
		commands := stateHandler.stateTransition(
			context.Background(),
			*provider,
			*providerSvc,
			st.Experiment,
			st.LastTransitionTime,
		)

		transitionTime := metav1.Time{}
		for _, command := range commands {
			if stat, ok := command.(SetStatus); ok {
				fmt.Println("found a setStatus yeeeeh")
				transitionTime = stat.LastTransitionTime
			}
		}

		for i, stCommand := range st.Commands {
			if stat, ok := stCommand.(SetStatus); ok {
				fmt.Println("found a setStatus on stCommands yeeeeh")
				stat.WithLastTransitionTime(transitionTime).statusWithCondition()
				fmt.Println("I've set the stat to ", stat)
				st.Commands[i] = stat
				fmt.Println("I've set the stCommands to ", st.Commands[i])
			}
		}

		Expect(commands).To(ContainElements(st.Commands))
		if st.workflowFactory.CalledWithProvider != nil {
			Expect(st.workflowFactory.CalledWithProvider).To(BeComparableTo(provider))
		}
		if st.workflowFactory.CalledWithProviderSvc != nil {
			Expect(st.workflowFactory.CalledWithProviderSvc).To(BeComparableTo(providerSvc))
		}
	},
		Check("Empty",
			From(UnknownState, emptyProviderId, "", v1, transitionTime).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithLastTransitionTime(transitionTime).
					WithSynchronizationState(apis.Creating).
					WithVersion(v1).
					statusWithCondition()).
				IssuesCreationWorkflow(),
		),
		Check("Empty and workflow creation fails",
			From(UnknownState, emptyProviderId, "", v1, transitionTime).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v1).
					WithLastTransitionTime(transitionTime).
					WithMessage(workflowconstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed).statusWithCondition()),
		),
		Check("Empty with version",
			From(UnknownState, emptyProviderId, v1, v1, transitionTime).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1).WithLastTransitionTime(transitionTime).statusWithCondition()).
				IssuesCreationWorkflow(),
		),
		Check("Empty with id",
			From(UnknownState, providerId, "", v1, transitionTime).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithProvider(providerId).
					WithVersion(v1).WithLastTransitionTime(transitionTime).statusWithCondition()).
				IssuesUpdateWorkflow(),
		),
		Check("Empty with id and workflow creation fails",
			From(UnknownState, providerId, "", v1, transitionTime).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithProvider(providerId).
					WithVersion(v1).WithLastTransitionTime(transitionTime).
					WithMessage(workflowconstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed).statusWithCondition()),
		),
		Check("Empty with id and version",
			From(UnknownState, providerId, v1, v2, transitionTime).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithProvider(providerId).
					WithLastTransitionTime(transitionTime).
					WithVersion(v2).statusWithCondition()).
				IssuesUpdateWorkflow(),
		),
		Check("Creating succeeds",
			From(apis.Creating, emptyProviderId, v1, irrelevant, transitionTime).
				AcquireExperiment().
				WithSucceededCreateWorkFlow(*provider, providerId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Succeeded).
					WithProvider(providerId).
					WithVersion(v1).WithLastTransitionTime(transitionTime).statusWithCondition()).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Creating succeeds without providerId or provider error",
			From(apis.Creating, providerId, v1, irrelevant, transitionTime).
				AcquireExperiment().
				WithSucceededCreateWorkFlow(*provider, emptyProviderId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProvider(providerId).
					WithVersion(v1).WithLastTransitionTime(transitionTime).
					WithMessage("id was empty").statusWithCondition()).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Creating succeeds with provider error",
			From(apis.Creating, anotherIdSameProvider, v1, irrelevant, transitionTime).
				AcquireExperiment().
				WithSucceededCreateWorkFlow(*provider, providerId, providerError).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProvider(providerId).
					WithVersion(v1).WithLastTransitionTime(transitionTime).
					WithMessage(providerError).statusWithCondition()).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Creating fails",
			From(apis.Creating, emptyProviderId, v1, irrelevant, transitionTime).
				AcquireExperiment().
				WithCreateWorkFlow(argo.WorkflowFailed).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithVersion(v1).WithLastTransitionTime(transitionTime).
					WithMessage("operation failed").statusWithCondition()).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Creating without version",
			From(apis.Creating, emptyProviderId, "", irrelevant, transitionTime).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithLastTransitionTime(transitionTime).
					WithMessage("creating resource with empty version").statusWithCondition()),
		),
		Check("Succeeded no update",
			From(apis.Succeeded, providerId, v1, v1, transitionTime).
				AcquireExperiment(),
		),
		Check("Succeeded with update",
			From(apis.Succeeded, providerId, v1, v2, transitionTime).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithProvider(providerId).
					WithLastTransitionTime(transitionTime).
					WithVersion(v2).statusWithCondition()).
				IssuesUpdateWorkflow(),
		),
		Check("Succeeded with update and workflow creation fails",
			From(apis.Succeeded, providerId, v1, v2, transitionTime).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithProvider(providerId).
					WithVersion(v2).
					WithLastTransitionTime(transitionTime).
					WithMessage(workflowconstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed).statusWithCondition()),
		),
		Check("Succeeded with update but no ProviderAndId",
			From(apis.Succeeded, emptyProviderId, v1, v2, transitionTime).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithLastTransitionTime(transitionTime).
					WithVersion(v2).statusWithCondition()).
				IssuesCreationWorkflow(),
		),
		Check("Succeeded with update and workflow creation fails",
			From(apis.Succeeded, emptyProviderId, v1, v2, transitionTime).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v2).
					WithLastTransitionTime(transitionTime).
					WithMessage(workflowconstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed).statusWithCondition()),
		),
		Check("Succeeded with update but no ProviderAndId and no version",
			From(apis.Succeeded, emptyProviderId, "", v1, transitionTime).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1).WithLastTransitionTime(transitionTime).statusWithCondition()).
				IssuesCreationWorkflow(),
		),
		Check("Failed no update",
			From(apis.Failed, providerId, v1, v1, transitionTime).
				AcquireExperiment(),
		),
		Check("Failed with update",
			From(apis.Failed, providerId, v1, v2, transitionTime).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithProvider(providerId).
					WithLastTransitionTime(transitionTime).
					WithVersion(v2).statusWithCondition()).
				IssuesUpdateWorkflow(),
		),
		Check("Failed with update and workflow creation fails",
			From(apis.Failed, providerId, v1, v2, transitionTime).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithProvider(providerId).
					WithVersion(v2).
					WithMessage(workflowconstants.ConstructionFailedError).
					WithLastTransitionTime(transitionTime).
					WithSynchronizationState(apis.Failed).statusWithCondition()),
		),
		Check("Failed with update but no ProviderAndId",
			From(apis.Failed, emptyProviderId, v1, v2, transitionTime).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithLastTransitionTime(transitionTime).
					WithVersion(v2).statusWithCondition()).
				IssuesCreationWorkflow(),
		),
		Check("Failed with update but no ProviderAndId and workflow creation fails",
			From(apis.Failed, emptyProviderId, v1, v2, transitionTime).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v2).
					WithMessage(workflowconstants.ConstructionFailedError).
					WithLastTransitionTime(transitionTime).
					WithSynchronizationState(apis.Failed).statusWithCondition()),
		),
		Check("Failed with update but no ProviderAndId and no version",
			From(apis.Failed, emptyProviderId, "", v1, transitionTime).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1).WithLastTransitionTime(transitionTime).statusWithCondition()).
				IssuesCreationWorkflow(),
		),
		Check("Updating succeeds with providerId",
			From(apis.Updating, providerId, v1, irrelevant, transitionTime).
				AcquireExperiment().
				WithSucceededUpdateWorkflow(*provider, anotherIdSameProvider, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Succeeded).
					WithProvider(anotherIdSameProvider).
					WithVersion(v1).WithLastTransitionTime(transitionTime).statusWithCondition()).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Updating succeeds without providerId or provider error",
			From(apis.Updating, providerId, v1, irrelevant, transitionTime).
				AcquireExperiment().
				WithSucceededUpdateWorkflow(*provider, emptyProviderId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProvider(providerId).
					WithVersion(v1).WithLastTransitionTime(transitionTime).
					WithMessage("id was empty").statusWithCondition()).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Updating succeeds with provider error",
			From(apis.Updating, providerId, v1, irrelevant, transitionTime).
				AcquireExperiment().
				WithSucceededUpdateWorkflow(*provider, anotherIdSameProvider, providerError).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProvider(anotherIdSameProvider).
					WithVersion(v1).WithLastTransitionTime(transitionTime).
					WithMessage(providerError).statusWithCondition()).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Updating fails",
			From(apis.Updating, providerId, v1, irrelevant, transitionTime).
				AcquireExperiment().
				WithFailedUpdateWorkflow().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProvider(providerId).
					WithVersion(v1).WithLastTransitionTime(transitionTime).
					WithMessage("operation failed").statusWithCondition()).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Updating without version",
			From(apis.Updating, providerId, "", irrelevant, transitionTime).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithProvider(providerId).
					WithLastTransitionTime(transitionTime).
					WithMessage("updating resource with empty version or providerId").statusWithCondition()),
		),
		Check("Updating without ProviderAndId",
			From(apis.Updating, emptyProviderId, v1, irrelevant, transitionTime).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithVersion(v1).WithLastTransitionTime(transitionTime).
					WithMessage("updating resource with empty version or providerId").statusWithCondition()),
		),
		Check("Updating without ProviderAndId or version",
			From(apis.Updating, emptyProviderId, "", irrelevant, transitionTime).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithLastTransitionTime(transitionTime).
					WithMessage("updating resource with empty version or providerId").statusWithCondition()),
		),
		Check("Deleting from Succeeded",
			From(apis.Succeeded, providerId, v1, irrelevant, transitionTime).
				AcquireExperiment().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithProvider(providerId).
					WithVersion(v1).WithLastTransitionTime(transitionTime).statusWithCondition()).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Succeeded without providerId",
			From(apis.Succeeded, emptyProviderId, v1, irrelevant, transitionTime).
				AcquireExperiment().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleted).
					WithVersion(v1).WithLastTransitionTime(transitionTime).statusWithCondition()),
		),
		Check("Deleting from Failed",
			From(apis.Failed, providerId, v1, irrelevant, transitionTime).
				AcquireExperiment().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithProvider(providerId).
					WithVersion(v1).WithLastTransitionTime(transitionTime).statusWithCondition()).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Failed without providerId",
			From(apis.Failed, emptyProviderId, v1, irrelevant, transitionTime).
				AcquireExperiment().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleted).
					WithVersion(v1).WithLastTransitionTime(transitionTime).statusWithCondition()),
		),
		Check("Deletion succeeds with providerId",
			From(apis.Deleting, providerId, v1, irrelevant, transitionTime).
				AcquireExperiment().
				DeletionRequested().
				WithSucceededDeletionWorkflow(*provider, anotherIdSameProvider, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithProvider(providerId).
					WithVersion(v1).WithLastTransitionTime(transitionTime).
					WithMessage("id should be empty").statusWithCondition()).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Deletion succeeds",
			From(apis.Deleting, providerId, v1, irrelevant, transitionTime).
				AcquireExperiment().
				DeletionRequested().
				WithSucceededDeletionWorkflow(*emptyProvider, emptyProviderId, "").
				IssuesCommand(*NewSetStatus().
					WithLastTransitionTime(transitionTime).
					WithSynchronizationState(apis.Deleted).
					WithProvider(emptyProviderId).
					WithVersion(v1).WithLastTransitionTime(transitionTime).statusWithCondition()).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Deletion succeeds with provider error",
			From(apis.Deleting, providerId, v1, irrelevant, transitionTime).
				AcquireExperiment().
				DeletionRequested().
				WithSucceededDeletionWorkflow(*provider, providerId, providerError).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithProvider(providerId).
					WithVersion(v1).WithLastTransitionTime(transitionTime).
					WithMessage(providerError).statusWithCondition()).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Deletion fails",
			From(apis.Deleting, providerId, v1, irrelevant, transitionTime).
				AcquireExperiment().
				DeletionRequested().
				WithDeletionWorkflow(argo.WorkflowFailed).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithProvider(providerId).
					WithVersion(v1).WithLastTransitionTime(transitionTime).
					WithMessage("operation failed").statusWithCondition()).
				MarksAllWorkflowsAsProcessed(),
		),
		Check("Stay in deleted",
			From(apis.Deleted, providerId, v1, irrelevant, transitionTime).
				ReleaseExperiment(),
		),
		Check("Any non-deleted state with different provider",
			From(anyNonDeletedState(), anotherProviderId, irrelevant, irrelevant, transitionTime).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithVersion(irrelevant).
					WithProvider(anotherProviderId).
					WithMessage(StateHandlerConstants.ProviderChangedError).
					WithLastTransitionTime(transitionTime).
					WithSynchronizationState(apis.Failed).statusWithCondition()),
		))
})
