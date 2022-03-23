//go:build unit
// +build unit

package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type RunConfigurationStateTransitionTestCase struct {
	workflowFactory  RunConfigurationWorkflowFactory
	RunConfiguration *pipelinesv1.RunConfiguration
	SystemStatus     StubbedWorkflows
	Commands         []Command
}

func (st RunConfigurationStateTransitionTestCase) WithWorkFlow(workflow *argo.Workflow) RunConfigurationStateTransitionTestCase {
	st.SystemStatus.AddWorkflow(*workflow)
	return st
}

func (st RunConfigurationStateTransitionTestCase) WithCreateWorkFlow(phase argo.WorkflowPhase) RunConfigurationStateTransitionTestCase {
	return st.WithWorkFlow(CreateTestWorkflow(RunConfigurationWorkflowConstants.CreateOperationLabel, phase))
}

func (st RunConfigurationStateTransitionTestCase) WithCreateWorkFlowWithId(phase argo.WorkflowPhase, kfpId string) RunConfigurationStateTransitionTestCase {
	return st.WithWorkFlow(
		setWorkflowOutputs(
			CreateTestWorkflow(RunConfigurationWorkflowConstants.CreateOperationLabel, phase),
			[]argo.Parameter{
				{
					Name:  RunConfigurationWorkflowConstants.RunConfigurationIdParameterName,
					Value: argo.AnyStringPtr(kfpId),
				},
			},
		),
	)
}

func (st RunConfigurationStateTransitionTestCase) WithFailedUpdateWorkflow() RunConfigurationStateTransitionTestCase {
	return st.WithWorkFlow(
		CreateTestWorkflow(RunConfigurationWorkflowConstants.UpdateOperationLabel, argo.WorkflowFailed),
	)
}

func (st RunConfigurationStateTransitionTestCase) WithSucceededUpdateWorkflowWithId(kfpId string) RunConfigurationStateTransitionTestCase {
	return st.WithWorkFlow(
		setWorkflowOutputs(
			CreateTestWorkflow(RunConfigurationWorkflowConstants.UpdateOperationLabel, argo.WorkflowSucceeded),
			[]argo.Parameter{
				{
					Name:  RunConfigurationWorkflowConstants.RunConfigurationIdParameterName,
					Value: argo.AnyStringPtr(kfpId),
				},
			},
		),
	)
}

func (st RunConfigurationStateTransitionTestCase) WithDeletionWorkflow(phase argo.WorkflowPhase) RunConfigurationStateTransitionTestCase {
	return st.WithWorkFlow(
		CreateTestWorkflow(RunConfigurationWorkflowConstants.DeleteOperationLabel, phase),
	)
}

func (st RunConfigurationStateTransitionTestCase) IssuesCreationWorkflow() RunConfigurationStateTransitionTestCase {
	creationWorkflow, _ := st.workflowFactory.ConstructCreationWorkflow(context.Background(), st.RunConfiguration)
	return st.IssuesCommand(CreateWorkflow{Workflow: *creationWorkflow})
}

func (st RunConfigurationStateTransitionTestCase) IssuesUpdateWorkflow() RunConfigurationStateTransitionTestCase {
	updateWorkflow, _ := st.workflowFactory.ConstructUpdateWorkflow(context.Background(), st.RunConfiguration)
	return st.IssuesCommand(CreateWorkflow{Workflow: *updateWorkflow})
}

func (st RunConfigurationStateTransitionTestCase) IssuesDeletionWorkflow() RunConfigurationStateTransitionTestCase {
	deletionWorkflow, _ := st.workflowFactory.ConstructDeletionWorkflow(context.Background(), st.RunConfiguration)
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
		WorkflowFactory: WorkflowFactory{
			Config: configv1.Configuration{
				DefaultExperiment: "Default",
				Argo: configv1.ArgoConfiguration{
					KfpSdkImage:   "kfp-sdk",
					CompilerImage: "compiler",
					ContainerDefaults: apiv1.Container{
						ImagePullPolicy: "Never",
					},
				},
				KfpEndpoint: "http://www.example.com",
			},
		},
	}

	kfpId := "12345"
	anotherKfpId := "67890"
	specv1 := RandomRunConfigurationSpec()
	v0 := pipelinesv1.RunConfigurationSpec{}.ComputeVersion()
	v1 := specv1.ComputeVersion()
	UnknownState := pipelinesv1.SynchronizationState(RandomString())

	var Check = func(description string, transition RunConfigurationStateTransitionTestCase) TableEntry {
		return Entry(
			description,
			transition,
		)
	}

	var From = func(status pipelinesv1.SynchronizationState, id string, version string) RunConfigurationStateTransitionTestCase {
		runConfiguration := RandomRunConfiguration()
		runConfiguration.Spec = specv1
		runConfiguration.Status = pipelinesv1.Status{
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
			WorkflowFactory:    workflowFactory,
		}
		commands := stateHandler.StateTransition(context.Background(), st.RunConfiguration)
		Expect(commands).To(Equal(st.Commands))
	},
		Check("Empty",
			From(UnknownState, "", "").
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Empty with version",
			From(UnknownState, "", v1).
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Empty with id",
			From(UnknownState, kfpId, "").
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Empty with id and version",
			From(UnknownState, kfpId, v1).
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Creating succeeds",
			From(pipelinesv1.Creating, "", v1).
				AcquireRunConfiguration().
				WithCreateWorkFlowWithId(argo.WorkflowSucceeded, kfpId).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Succeeded).
					WithKfpId(kfpId).
					WithVersion(v1)).
				DeletesAllWorkflows(),
		),
		Check("Creating succeeds with existing KfpId",
			From(pipelinesv1.Creating, anotherKfpId, v1).
				AcquireRunConfiguration().
				WithCreateWorkFlowWithId(argo.WorkflowSucceeded, kfpId).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Succeeded).
					WithKfpId(kfpId).
					WithVersion(v1)).
				DeletesAllWorkflows(),
		),
		Check("Creating fails",
			From(pipelinesv1.Creating, "", v1).
				AcquireRunConfiguration().
				WithCreateWorkFlow(argo.WorkflowFailed).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Failed).
					WithVersion(v1).
					WithMessage("run configuration creation failed")).
				DeletesAllWorkflows(),
		),
		Check("Creating without version",
			From(pipelinesv1.Creating, "", "").
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Failed).
					WithMessage("creating run configuration with empty version")),
		),
		Check("Succeeded no update",
			From(pipelinesv1.Succeeded, kfpId, v1).
				AcquireRunConfiguration(),
		),
		Check("Succeeded with update",
			From(pipelinesv1.Succeeded, kfpId, v0).
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Succeeded with update but no KfpId",
			From(pipelinesv1.Succeeded, "", v0).
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Succeeded with update but no KfpId and no version",
			From(pipelinesv1.Succeeded, "", "").
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Failed no update",
			From(pipelinesv1.Failed, kfpId, v1).
				AcquireRunConfiguration(),
		),
		Check("Failed with Update",
			From(pipelinesv1.Failed, kfpId, v0).
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Failed with Update but no KfpId",
			From(pipelinesv1.Failed, "", v0).
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Failed with Update but no KfpId and no version",
			From(pipelinesv1.Failed, "", "").
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Updating succeeds with kfpId",
			From(pipelinesv1.Updating, anotherKfpId, v1).
				AcquireRunConfiguration().
				WithSucceededUpdateWorkflowWithId(kfpId).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Succeeded).
					WithKfpId(kfpId).
					WithVersion(v1)).
				DeletesAllWorkflows(),
		),
		Check("Updating succeeds without kfpId",
			From(pipelinesv1.Updating, anotherKfpId, v1).
				AcquireRunConfiguration().
				WithSucceededUpdateWorkflowWithId("").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Failed).
					WithVersion(v1).
					WithMessage("could not retrieve kfpId")).
				DeletesAllWorkflows(),
		),
		Check("Updating fails",
			From(pipelinesv1.Updating, kfpId, v1).
				AcquireRunConfiguration().
				WithFailedUpdateWorkflow().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Failed).
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage("run configuration update failed")).
				DeletesAllWorkflows(),
		),
		Check("Updating without version",
			From(pipelinesv1.Updating, kfpId, "").
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Failed).
					WithKfpId(kfpId).
					WithMessage("updating run configuration with empty version or kfpId")),
		),
		Check("Updating without KfpId",
			From(pipelinesv1.Updating, "", v1).
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Failed).
					WithVersion(v1).
					WithMessage("updating run configuration with empty version or kfpId")),
		),
		Check("Updating without KfpId or version",
			From(pipelinesv1.Updating, "", "").
				AcquireRunConfiguration().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Failed).
					WithMessage("updating run configuration with empty version or kfpId")),
		),
		Check("Deleting from Succeeded",
			From(pipelinesv1.Succeeded, kfpId, v1).
				AcquireRunConfiguration().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Deleting).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Succeeded without kfpId",
			From(pipelinesv1.Succeeded, "", v1).
				AcquireRunConfiguration().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Deleted).
					WithVersion(v1)),
		),
		Check("Deleting from Failed",
			From(pipelinesv1.Failed, kfpId, v1).
				AcquireRunConfiguration().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Deleting).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Failed without kfpId",
			From(pipelinesv1.Failed, "", v1).
				AcquireRunConfiguration().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Deleted).
					WithVersion(v1)),
		),
		Check("Deletion succeeds",
			From(pipelinesv1.Deleting, kfpId, v1).
				AcquireRunConfiguration().
				DeletionRequested().
				WithDeletionWorkflow(argo.WorkflowSucceeded).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Deleted).
					WithKfpId(kfpId).
					WithVersion(v1)).
				DeletesAllWorkflows(),
		),
		Check("Deletion fails",
			From(pipelinesv1.Deleting, kfpId, v1).
				AcquireRunConfiguration().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(pipelinesv1.Deleting).
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage("run configuration deletion failed")).
				WithDeletionWorkflow(argo.WorkflowFailed).
				DeletesAllWorkflows(),
		),
		Check("Stay in deleted",
			From(pipelinesv1.Deleted, kfpId, v1).
				ReleaseRunConfiguration(),
		))
})
