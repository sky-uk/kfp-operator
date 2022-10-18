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
	providers "github.com/sky-uk/kfp-operator/providers/base"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type StateTransitionTestCase struct {
	workflowFactory WorkflowFactory[*apis.TestResource]
	Experiment      *apis.TestResource
	SystemStatus    StubbedWorkflows
	Commands        []Command
}

type SucceedingWorkflowFactory struct{}

func (f SucceedingWorkflowFactory) ConstructCreationWorkflow(_ *apis.TestResource) (*argo.Workflow, error) {
	return &argo.Workflow{}, nil
}

func (f SucceedingWorkflowFactory) ConstructUpdateWorkflow(_ *apis.TestResource) (*argo.Workflow, error) {
	return &argo.Workflow{}, nil
}

func (f SucceedingWorkflowFactory) ConstructDeletionWorkflow(_ *apis.TestResource) (*argo.Workflow, error) {
	return &argo.Workflow{}, nil
}

type FailingWorkflowFactory struct{}

func (f FailingWorkflowFactory) ConstructCreationWorkflow(_ *apis.TestResource) (*argo.Workflow, error) {
	return nil, fmt.Errorf("an error occurred")
}

func (f FailingWorkflowFactory) ConstructUpdateWorkflow(_ *apis.TestResource) (*argo.Workflow, error) {
	return nil, fmt.Errorf("an error occurred")
}

func (f FailingWorkflowFactory) ConstructDeletionWorkflow(_ *apis.TestResource) (*argo.Workflow, error) {
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

func (st StateTransitionTestCase) WithSucceededCreateWorkFlow(kfpId string, providerError string) StateTransitionTestCase {
	return st.WithWorkFlow(
		setProviderOutput(
			CreateTestWorkflow(WorkflowConstants.CreateOperationLabel, argo.WorkflowSucceeded),
			providers.Output{Id: kfpId, ProviderError: providerError},
		),
	)
}

func (st StateTransitionTestCase) WithFailedUpdateWorkflow() StateTransitionTestCase {
	return st.WithWorkFlow(
		CreateTestWorkflow(WorkflowConstants.UpdateOperationLabel, argo.WorkflowFailed),
	)
}

func (st StateTransitionTestCase) WithSucceededUpdateWorkflow(kfpId string, providerError string) StateTransitionTestCase {
	return st.WithWorkFlow(
		setProviderOutput(
			CreateTestWorkflow(WorkflowConstants.UpdateOperationLabel, argo.WorkflowSucceeded),
			providers.Output{Id: kfpId, ProviderError: providerError},
		),
	)
}

func (st StateTransitionTestCase) WithDeletionWorkflow(phase argo.WorkflowPhase) StateTransitionTestCase {
	return st.WithWorkFlow(
		CreateTestWorkflow(WorkflowConstants.DeleteOperationLabel, phase),
	)
}

func (st StateTransitionTestCase) WithSucceededDeletionWorkflow(kfpId string, providerError string) StateTransitionTestCase {
	return st.WithWorkFlow(
		setProviderOutput(
			CreateTestWorkflow(WorkflowConstants.DeleteOperationLabel, argo.WorkflowSucceeded),
			providers.Output{Id: kfpId, ProviderError: providerError},
		),
	)
}

func (st StateTransitionTestCase) IssuesCreationWorkflow() StateTransitionTestCase {
	creationWorkflow, _ := st.workflowFactory.ConstructCreationWorkflow(st.Experiment)
	return st.IssuesCommand(CreateWorkflow{Workflow: *creationWorkflow})
}

func (st StateTransitionTestCase) IssuesUpdateWorkflow() StateTransitionTestCase {
	updateWorkflow, _ := st.workflowFactory.ConstructUpdateWorkflow(st.Experiment)
	return st.IssuesCommand(CreateWorkflow{Workflow: *updateWorkflow})
}

func (st StateTransitionTestCase) IssuesDeletionWorkflow() StateTransitionTestCase {
	deletionWorkflow, _ := st.workflowFactory.ConstructDeletionWorkflow(st.Experiment)
	return st.IssuesCommand(CreateWorkflow{Workflow: *deletionWorkflow})
}

func (st StateTransitionTestCase) DeletesAllWorkflows() StateTransitionTestCase {
	return st.IssuesCommand(DeleteWorkflows{
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

var _ = Describe("Experiment State handler", func() {
	kfpId := "12345"
	anotherKfpId := "67890"
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

	var From = func(status apis.SynchronizationState, id string, versionInState string, computedVersion string) StateTransitionTestCase {
		resource := apis.RandomResource()
		resource.SetStatus(apis.Status{
			SynchronizationState: status,
			Version:              versionInState,
			KfpId:                id,
		})
		resource.SetComputedVersion(computedVersion)

		return StateTransitionTestCase{
			workflowFactory: SucceedingWorkflowFactory{}, // TODO: mock workflowFactory
			Experiment:      resource,
			Commands:        []Command{},
		}
	}

	DescribeTable("State transitions", func(st StateTransitionTestCase) {
		var stateHandler = StateHandler[*apis.TestResource]{
			WorkflowRepository: st.SystemStatus,
			WorkflowFactory:    st.workflowFactory,
		}
		commands := stateHandler.stateTransition(context.Background(), st.Experiment)
		Expect(commands).To(Equal(st.Commands))
	},
		Check("Empty",
			From(UnknownState, "", "", v1).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Empty and workflow creation fails",
			From(UnknownState, "", "", v1).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Empty with version",
			From(UnknownState, "", v1, v1).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Empty with id",
			From(UnknownState, kfpId, "", v1).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Empty with id and workflow creation fails",
			From(UnknownState, kfpId, "", v1).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Empty with id and version",
			From(UnknownState, kfpId, v1, v2).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithKfpId(kfpId).
					WithVersion(v2)).
				IssuesUpdateWorkflow(),
		),
		Check("Creating succeeds",
			From(apis.Creating, "", v1, irrelevant).
				AcquireExperiment().
				WithSucceededCreateWorkFlow(kfpId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Succeeded).
					WithKfpId(kfpId).
					WithVersion(v1)).
				DeletesAllWorkflows(),
		),
		Check("Creating succeeds without kfpId or provider error",
			From(apis.Creating, kfpId, v1, irrelevant).
				AcquireExperiment().
				WithSucceededCreateWorkFlow("", "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage("id was empty")).
				DeletesAllWorkflows(),
		),
		Check("Creating succeeds with provider error",
			From(apis.Creating, anotherKfpId, v1, irrelevant).
				AcquireExperiment().
				WithSucceededCreateWorkFlow(kfpId, providerError).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage(providerError)).
				DeletesAllWorkflows(),
		),
		Check("Creating fails",
			From(apis.Creating, "", v1, irrelevant).
				AcquireExperiment().
				WithCreateWorkFlow(argo.WorkflowFailed).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithVersion(v1).
					WithMessage("operation failed")).
				DeletesAllWorkflows(),
		),
		Check("Creating without version",
			From(apis.Creating, "", "", irrelevant).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithMessage("creating resource with empty version")),
		),
		Check("Succeeded no update",
			From(apis.Succeeded, kfpId, v1, v1).
				AcquireExperiment(),
		),
		Check("Succeeded with update",
			From(apis.Succeeded, kfpId, v1, v2).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithKfpId(kfpId).
					WithVersion(v2)).
				IssuesUpdateWorkflow(),
		),
		Check("Succeeded with update and workflow creation fails",
			From(apis.Succeeded, kfpId, v1, v2).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithKfpId(kfpId).
					WithVersion(v2).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Succeeded with update but no KfpId",
			From(apis.Succeeded, "", v1, v2).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v2)).
				IssuesCreationWorkflow(),
		),
		Check("Succeeded with update and workflow creation fails",
			From(apis.Succeeded, "", v1, v2).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v2).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Succeeded with update but no KfpId and no version",
			From(apis.Succeeded, "", "", v1).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Failed no update",
			From(apis.Failed, kfpId, v1, v1).
				AcquireExperiment(),
		),
		Check("Failed with update",
			From(apis.Failed, kfpId, v1, v2).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithKfpId(kfpId).
					WithVersion(v2)).
				IssuesUpdateWorkflow(),
		),
		Check("Failed with update and workflow creation fails",
			From(apis.Failed, kfpId, v1, v2).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithKfpId(kfpId).
					WithVersion(v2).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Failed with update but no KfpId",
			From(apis.Failed, "", v1, v2).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v2)).
				IssuesCreationWorkflow(),
		),
		Check("Failed with update but no KfpId and workflow creation fails",
			From(apis.Failed, "", v1, v2).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v2).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Failed with update but no KfpId and no version",
			From(apis.Failed, "", "", v1).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Updating succeeds with kfpId",
			From(apis.Updating, kfpId, v1, irrelevant).
				AcquireExperiment().
				WithSucceededUpdateWorkflow(anotherKfpId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Succeeded).
					WithKfpId(anotherKfpId).
					WithVersion(v1)).
				DeletesAllWorkflows(),
		),
		Check("Updating succeeds without kfpId or provider error",
			From(apis.Updating, kfpId, v1, irrelevant).
				AcquireExperiment().
				WithSucceededUpdateWorkflow("", "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage("id was empty")).
				DeletesAllWorkflows(),
		),
		Check("Updating succeeds with provider error",
			From(apis.Updating, kfpId, v1, irrelevant).
				AcquireExperiment().
				WithSucceededUpdateWorkflow(anotherKfpId, providerError).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithKfpId(anotherKfpId).
					WithVersion(v1).
					WithMessage(providerError)).
				DeletesAllWorkflows(),
		),
		Check("Updating fails",
			From(apis.Updating, kfpId, v1, irrelevant).
				AcquireExperiment().
				WithFailedUpdateWorkflow().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage("operation failed")).
				DeletesAllWorkflows(),
		),
		Check("Updating without version",
			From(apis.Updating, kfpId, "", irrelevant).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithKfpId(kfpId).
					WithMessage("updating resource with empty version or kfpId")),
		),
		Check("Updating without KfpId",
			From(apis.Updating, "", v1, irrelevant).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithVersion(v1).
					WithMessage("updating resource with empty version or kfpId")),
		),
		Check("Updating without KfpId or version",
			From(apis.Updating, "", "", irrelevant).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithMessage("updating resource with empty version or kfpId")),
		),
		Check("Deleting from Succeeded",
			From(apis.Succeeded, kfpId, v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Succeeded without kfpId",
			From(apis.Succeeded, "", v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleted).
					WithVersion(v1)),
		),
		Check("Deleting from Failed",
			From(apis.Failed, kfpId, v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Failed without kfpId",
			From(apis.Failed, "", v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleted).
					WithVersion(v1)),
		),
		Check("Deletion succeeds with kfpId",
			From(apis.Deleting, kfpId, v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				WithSucceededDeletionWorkflow(anotherKfpId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage("id should be empty")).
				DeletesAllWorkflows(),
		),
		Check("Deletion succeeds",
			From(apis.Deleting, kfpId, v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				WithSucceededDeletionWorkflow("", "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleted).
					WithKfpId("").
					WithVersion(v1)).
				DeletesAllWorkflows(),
		),
		Check("Deletion succeeds with provider error",
			From(apis.Deleting, kfpId, v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				WithSucceededDeletionWorkflow(kfpId, providerError).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage(providerError)).
				DeletesAllWorkflows(),
		),
		Check("Deletion fails",
			From(apis.Deleting, kfpId, v1, irrelevant).
				AcquireExperiment().
				DeletionRequested().
				WithDeletionWorkflow(argo.WorkflowFailed).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage("operation failed")).
				DeletesAllWorkflows(),
		),
		Check("Stay in deleted",
			From(apis.Deleted, kfpId, v1, irrelevant).
				ReleaseExperiment(),
		))
})
