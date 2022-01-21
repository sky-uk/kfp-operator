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

func (st RunConfigurationStateTransitionTestCase) To(state pipelinesv1.SynchronizationState, id string, version string) RunConfigurationStateTransitionTestCase {
	return st.IssuesCommand(SetStatus{Status: pipelinesv1.Status{
		KfpId:                id,
		Version:              version,
		SynchronizationState: state,
	}})
}

func (st RunConfigurationStateTransitionTestCase) WithWorkFlow(workflow *argo.Workflow) RunConfigurationStateTransitionTestCase {
	st.SystemStatus.AddWorkflow(*workflow)
	return st
}

func (st RunConfigurationStateTransitionTestCase) WithCreateWorkFlow(phase argo.WorkflowPhase) RunConfigurationStateTransitionTestCase {
	return st.WithWorkFlow(st.SystemStatus.CreateWorkflow(RunConfigurationWorkflowConstants.CreateOperationLabel, phase))
}

func (st RunConfigurationStateTransitionTestCase) WithCreateWorkFlowWithId(phase argo.WorkflowPhase, kfpId string) RunConfigurationStateTransitionTestCase {
	return st.WithWorkFlow(
		setWorkflowOutputs(
			st.SystemStatus.CreateWorkflow(RunConfigurationWorkflowConstants.CreateOperationLabel, phase),
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
		st.SystemStatus.CreateWorkflow(RunConfigurationWorkflowConstants.UpdateOperationLabel, argo.WorkflowFailed),
	)
}

func (st RunConfigurationStateTransitionTestCase) WithSucceededUpdateWorkflowWithId(kfpId string) RunConfigurationStateTransitionTestCase {
	return st.WithWorkFlow(
		setWorkflowOutputs(
			st.SystemStatus.CreateWorkflow(RunConfigurationWorkflowConstants.UpdateOperationLabel, argo.WorkflowSucceeded),
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
		st.SystemStatus.CreateWorkflow(RunConfigurationWorkflowConstants.DeleteOperationLabel, phase),
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
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Empty with version",
			From(UnknownState, "", v1).
				AcquireRunConfiguration().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Empty with id",
			From(UnknownState, kfpId, "").
				AcquireRunConfiguration().
				To(pipelinesv1.Updating, kfpId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Empty with id and version",
			From(UnknownState, kfpId, v1).
				AcquireRunConfiguration().
				To(pipelinesv1.Updating, kfpId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Creating succeeds",
			From(pipelinesv1.Creating, "", v1).
				AcquireRunConfiguration().
				WithCreateWorkFlowWithId(argo.WorkflowSucceeded, kfpId).
				To(pipelinesv1.Succeeded, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Creating succeeds with existing KfpId",
			From(pipelinesv1.Creating, anotherKfpId, v1).
				AcquireRunConfiguration().
				WithCreateWorkFlowWithId(argo.WorkflowSucceeded, kfpId).
				To(pipelinesv1.Succeeded, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Creating fails",
			From(pipelinesv1.Creating, "", v1).
				AcquireRunConfiguration().
				WithCreateWorkFlow(argo.WorkflowFailed).
				To(pipelinesv1.Failed, "", v1).
				DeletesAllWorkflows(),
		),
		Check("Creating without version",
			From(pipelinesv1.Creating, "", "").
				AcquireRunConfiguration().
				To(pipelinesv1.Failed, "", ""),
		),
		Check("Succeeded no update",
			From(pipelinesv1.Succeeded, kfpId, v1).
				AcquireRunConfiguration(),
		),
		Check("Succeeded with update",
			From(pipelinesv1.Succeeded, kfpId, v0).
				AcquireRunConfiguration().
				To(pipelinesv1.Updating, kfpId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Succeeded with update but no KfpId",
			From(pipelinesv1.Succeeded, "", v0).
				AcquireRunConfiguration().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Succeeded with update but no KfpId and no version",
			From(pipelinesv1.Succeeded, "", "").
				AcquireRunConfiguration().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Failed no update",
			From(pipelinesv1.Failed, kfpId, v1).
				AcquireRunConfiguration(),
		),
		Check("Failed with Update",
			From(pipelinesv1.Failed, kfpId, v0).
				AcquireRunConfiguration().
				To(pipelinesv1.Updating, kfpId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Failed with Update but no KfpId",
			From(pipelinesv1.Failed, "", v0).
				AcquireRunConfiguration().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Failed with Update but no KfpId and no version",
			From(pipelinesv1.Failed, "", "").
				AcquireRunConfiguration().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Updating succeeds with kfpId",
			From(pipelinesv1.Updating, anotherKfpId, v1).
				AcquireRunConfiguration().
				WithSucceededUpdateWorkflowWithId(kfpId).
				To(pipelinesv1.Succeeded, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Updating succeeds without kfpId",
			From(pipelinesv1.Updating, anotherKfpId, v1).
				AcquireRunConfiguration().
				WithSucceededUpdateWorkflowWithId("").
				To(pipelinesv1.Failed, "", v1).
				DeletesAllWorkflows(),
		),
		Check("Updating fails",
			From(pipelinesv1.Updating, kfpId, v1).
				AcquireRunConfiguration().
				WithFailedUpdateWorkflow().
				To(pipelinesv1.Failed, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Updating without version",
			From(pipelinesv1.Updating, kfpId, "").
				AcquireRunConfiguration().
				To(pipelinesv1.Failed, kfpId, ""),
		),
		Check("Updating without KfpId",
			From(pipelinesv1.Updating, "", v1).
				AcquireRunConfiguration().
				To(pipelinesv1.Failed, "", v1),
		),
		Check("Updating without KfpId or version",
			From(pipelinesv1.Updating, "", "").
				AcquireRunConfiguration().
				To(pipelinesv1.Failed, "", ""),
		),
		Check("Deleting from Succeeded",
			From(pipelinesv1.Succeeded, kfpId, v1).
				AcquireRunConfiguration().
				DeletionRequested().
				To(pipelinesv1.Deleting, kfpId, v1).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Succeeded without kfpId",
			From(pipelinesv1.Succeeded, "", v1).
				AcquireRunConfiguration().
				DeletionRequested().
				To(pipelinesv1.Deleted, "", v1),
		),
		Check("Deleting from Failed",
			From(pipelinesv1.Failed, kfpId, v1).
				AcquireRunConfiguration().
				DeletionRequested().
				To(pipelinesv1.Deleting, kfpId, v1).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Failed without kfpId",
			From(pipelinesv1.Failed, "", v1).
				AcquireRunConfiguration().
				DeletionRequested().
				To(pipelinesv1.Deleted, "", v1),
		),
		Check("Deletion succeeds",
			From(pipelinesv1.Deleting, kfpId, v1).
				AcquireRunConfiguration().
				DeletionRequested().
				WithDeletionWorkflow(argo.WorkflowSucceeded).
				To(pipelinesv1.Deleted, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Deletion fails",
			From(pipelinesv1.Deleting, kfpId, v1).
				AcquireRunConfiguration().
				DeletionRequested().
				WithDeletionWorkflow(argo.WorkflowFailed).
				To(pipelinesv1.Deleting, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Stay in deleted",
			From(pipelinesv1.Deleted, kfpId, v1).
				ReleaseRunConfiguration(),
		))
})
