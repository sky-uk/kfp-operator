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
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha3"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	"github.com/sky-uk/kfp-operator/providers/base"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type RunConfigurationStateTransitionTestCase struct {
	workflowFactory  WorkflowFactory[*pipelinesv1.RunConfiguration]
	RunConfiguration *pipelinesv1.RunConfiguration
	SystemStatus     StubbedWorkflows
	Commands         []Command
}

type FailingRunConfigurationWorkflowFactory struct{}

func (f FailingRunConfigurationWorkflowFactory) ConstructCreationWorkflow(_ *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	return nil, fmt.Errorf("an error occurred")
}

func (f FailingRunConfigurationWorkflowFactory) ConstructUpdateWorkflow(_ *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	return nil, fmt.Errorf("an error occurred")
}

func (f FailingRunConfigurationWorkflowFactory) ConstructDeletionWorkflow(_ *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	return nil, fmt.Errorf("an error occurred")
}

func (st RunConfigurationStateTransitionTestCase) WorkflowConstructionFails() RunConfigurationStateTransitionTestCase {
	st.workflowFactory = FailingRunConfigurationWorkflowFactory{}
	return st
}

func (st RunConfigurationStateTransitionTestCase) WithWorkFlow(workflow *argo.Workflow) RunConfigurationStateTransitionTestCase {
	st.SystemStatus.AddWorkflow(*workflow)
	return st
}

func (st RunConfigurationStateTransitionTestCase) WithCreateWorkFlow(phase argo.WorkflowPhase) RunConfigurationStateTransitionTestCase {
	return st.WithWorkFlow(CreateTestWorkflow(WorkflowConstants.CreateOperationLabel, phase))
}

func (st RunConfigurationStateTransitionTestCase) WithCreateWorkFlowWithId(phase argo.WorkflowPhase, kfpId string) RunConfigurationStateTransitionTestCase {
	return st.WithWorkFlow(
		setProviderOutput(
			CreateTestWorkflow(WorkflowConstants.CreateOperationLabel, phase),
			base.Output{Id: kfpId},
		),
	)
}

func (st RunConfigurationStateTransitionTestCase) WithFailedUpdateWorkflow() RunConfigurationStateTransitionTestCase {
	return st.WithWorkFlow(
		CreateTestWorkflow(WorkflowConstants.UpdateOperationLabel, argo.WorkflowFailed),
	)
}

func (st RunConfigurationStateTransitionTestCase) WithSucceededUpdateWorkflowWithId(kfpId string) RunConfigurationStateTransitionTestCase {
	return st.WithWorkFlow(
		setProviderOutput(
			CreateTestWorkflow(WorkflowConstants.UpdateOperationLabel, argo.WorkflowSucceeded),
			base.Output{Id: kfpId},
		),
	)
}

func (st RunConfigurationStateTransitionTestCase) WithDeletionWorkflow(phase argo.WorkflowPhase) RunConfigurationStateTransitionTestCase {
	return st.WithWorkFlow(
		CreateTestWorkflow(WorkflowConstants.DeleteOperationLabel, phase),
	)
}

func (st RunConfigurationStateTransitionTestCase) IssuesCreationWorkflow() RunConfigurationStateTransitionTestCase {
	creationWorkflow, _ := st.workflowFactory.ConstructCreationWorkflow(st.RunConfiguration)
	return st.IssuesCommand(CreateWorkflow{Workflow: *creationWorkflow})
}

func (st RunConfigurationStateTransitionTestCase) IssuesUpdateWorkflow() RunConfigurationStateTransitionTestCase {
	updateWorkflow, _ := st.workflowFactory.ConstructUpdateWorkflow(st.RunConfiguration)
	return st.IssuesCommand(CreateWorkflow{Workflow: *updateWorkflow})
}

func (st RunConfigurationStateTransitionTestCase) IssuesDeletionWorkflow() RunConfigurationStateTransitionTestCase {
	deletionWorkflow, _ := st.workflowFactory.ConstructDeletionWorkflow(st.RunConfiguration)
	return st.IssuesCommand(CreateWorkflow{Workflow: *deletionWorkflow})
}

func (st RunConfigurationStateTransitionTestCase) DeletesAllWorkflows() RunConfigurationStateTransitionTestCase {
	return st.IssuesCommand(DeleteWorkflows{
		Workflows: st.SystemStatus.Workflows,
	})
}

func (st RunConfigurationStateTransitionTestCase) AcquireRunConfiguration() RunConfigurationStateTransitionTestCase {
	return st.IssuesCommand(AcquireResource{})
}

func (st RunConfigurationStateTransitionTestCase) ReleaseRunConfiguration() RunConfigurationStateTransitionTestCase {
	return st.IssuesCommand(ReleaseResource{})
}

func (st RunConfigurationStateTransitionTestCase) IssuesCommand(command Command) RunConfigurationStateTransitionTestCase {
	st.Commands = append(st.Commands, command)
	return st
}

func (st RunConfigurationStateTransitionTestCase) DeletionRequested() RunConfigurationStateTransitionTestCase {
	st.RunConfiguration.DeletionTimestamp = &metav1.Time{time.UnixMilli(1)}
	return st
}

var _ = Describe("RunConfiguration State handler", func() {
	// TODO: mock workflowFactory
	var workflowFactory = RunConfigurationWorkflowFactory{
		WorkflowFactoryBase: WorkflowFactoryBase{
			Config: config.Configuration{
				DefaultExperiment: "Default",
			},
		},
	}

	kfpId := "12345"
	anotherKfpId := "67890"
	rcv1 := pipelinesv1.RandomRunConfiguration()
	v0 := pipelinesv1.RunConfiguration{}.ComputeVersion()
	v1 := rcv1.ComputeVersion()
	UnknownState := apis.SynchronizationState(apis.RandomString())

	var Check = func(description string, transition RunConfigurationStateTransitionTestCase) TableEntry {
		return Entry(
			description,
			transition,
		)
	}

	var From = func(status apis.SynchronizationState, id string, version string) RunConfigurationStateTransitionTestCase {
		runConfiguration := rcv1.DeepCopy()
		runConfiguration.Status.Status = apis.Status{
			SynchronizationState: status,
			Version:              version,
			KfpId:                id,
		}

		return RunConfigurationStateTransitionTestCase{
			workflowFactory:  workflowFactory,
			RunConfiguration: runConfiguration,
			Commands:         []Command{},
		}
	}

	DescribeTable("State transitions", func(st RunConfigurationStateTransitionTestCase) {
		var stateHandler = RunConfigurationStateHandler{
			WorkflowRepository: st.SystemStatus,
			WorkflowFactory:    st.workflowFactory,
		}
		commands := stateHandler.StateTransition(context.Background(), st.RunConfiguration)
		Expect(commands).To(Equal(st.Commands))
	},
		Check("Empty",
			From(UnknownState, "", "").
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Empty and workflow creation fails",
			From(UnknownState, "", "").
				AcquireRunConfiguration().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Empty with version",
			From(UnknownState, "", v1).
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Empty with id",
			From(UnknownState, kfpId, "").
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Empty with id and workflow creation fails",
			From(UnknownState, kfpId, "").
				AcquireRunConfiguration().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Empty with id and version",
			From(UnknownState, kfpId, v1).
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Creating succeeds",
			From(apis.Creating, "", v1).
				AcquireRunConfiguration().
				WithCreateWorkFlowWithId(argo.WorkflowSucceeded, kfpId).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Succeeded).
					WithKfpId(kfpId).
					WithVersion(v1)).
				DeletesAllWorkflows(),
		),
		Check("Creating succeeds with existing KfpId",
			From(apis.Creating, anotherKfpId, v1).
				AcquireRunConfiguration().
				WithCreateWorkFlowWithId(argo.WorkflowSucceeded, kfpId).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Succeeded).
					WithKfpId(kfpId).
					WithVersion(v1)).
				DeletesAllWorkflows(),
		),
		Check("Creating fails",
			From(apis.Creating, "", v1).
				AcquireRunConfiguration().
				WithCreateWorkFlow(argo.WorkflowFailed).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithVersion(v1).
					WithMessage("run configuration creation failed")).
				DeletesAllWorkflows(),
		),
		Check("Creating without version",
			From(apis.Creating, "", "").
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithMessage("creating run configuration with empty version")),
		),
		Check("Succeeded no update",
			From(apis.Succeeded, kfpId, v1).
				AcquireRunConfiguration(),
		),
		Check("Succeeded with update",
			From(apis.Succeeded, kfpId, v0).
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Succeeded with update and workflow creation fails",
			From(apis.Succeeded, kfpId, v0).
				AcquireRunConfiguration().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Succeeded with update but no KfpId",
			From(apis.Succeeded, "", v0).
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Succeeded with update but no KfpId and workflow creation fails",
			From(apis.Succeeded, "", v0).
				AcquireRunConfiguration().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Succeeded with update but no KfpId and no version",
			From(apis.Succeeded, "", "").
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Failed no update",
			From(apis.Failed, kfpId, v1).
				AcquireRunConfiguration(),
		),
		Check("Failed with update",
			From(apis.Failed, kfpId, v0).
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Failed with update and workflow creation fails",
			From(apis.Failed, kfpId, v0).
				AcquireRunConfiguration().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Failed with update but no KfpId",
			From(apis.Failed, "", v0).
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Failed with update but no KfpId and workflow creation fails",
			From(apis.Failed, "", v0).
				AcquireRunConfiguration().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Failed with update but no KfpId and no version",
			From(apis.Failed, "", "").
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Updating succeeds with kfpId",
			From(apis.Updating, anotherKfpId, v1).
				AcquireRunConfiguration().
				WithSucceededUpdateWorkflowWithId(kfpId).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Succeeded).
					WithKfpId(kfpId).
					WithVersion(v1)).
				DeletesAllWorkflows(),
		),
		Check("Updating succeeds without kfpId",
			From(apis.Updating, anotherKfpId, v1).
				AcquireRunConfiguration().
				WithSucceededUpdateWorkflowWithId("").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithVersion(v1).
					WithMessage("could not retrieve kfpId")).
				DeletesAllWorkflows(),
		),
		Check("Updating fails",
			From(apis.Updating, kfpId, v1).
				AcquireRunConfiguration().
				WithFailedUpdateWorkflow().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage("run configuration update failed")).
				DeletesAllWorkflows(),
		),
		Check("Updating without version",
			From(apis.Updating, kfpId, "").
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithKfpId(kfpId).
					WithMessage("updating run configuration with empty version or kfpId")),
		),
		Check("Updating without KfpId",
			From(apis.Updating, "", v1).
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithVersion(v1).
					WithMessage("updating run configuration with empty version or kfpId")),
		),
		Check("Updating without KfpId or version",
			From(apis.Updating, "", "").
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithMessage("updating run configuration with empty version or kfpId")),
		),
		Check("Deleting from Succeeded",
			From(apis.Succeeded, kfpId, v1).
				AcquireRunConfiguration().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Succeeded without kfpId",
			From(apis.Succeeded, "", v1).
				AcquireRunConfiguration().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleted).
					WithVersion(v1)),
		),
		Check("Deleting from Failed",
			From(apis.Failed, kfpId, v1).
				AcquireRunConfiguration().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Failed without kfpId",
			From(apis.Failed, "", v1).
				AcquireRunConfiguration().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleted).
					WithVersion(v1)),
		),
		Check("Deletion succeeds",
			From(apis.Deleting, kfpId, v1).
				AcquireRunConfiguration().
				DeletionRequested().
				WithDeletionWorkflow(argo.WorkflowSucceeded).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleted).
					WithKfpId(kfpId).
					WithVersion(v1)).
				DeletesAllWorkflows(),
		),
		Check("Deletion fails",
			From(apis.Deleting, kfpId, v1).
				AcquireRunConfiguration().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage("run configuration deletion failed")).
				WithDeletionWorkflow(argo.WorkflowFailed).
				DeletesAllWorkflows(),
		),
		Check("Stay in deleted",
			From(apis.Deleted, kfpId, v1).
				ReleaseRunConfiguration(),
		))
})
