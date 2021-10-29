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
	Commands         []RunConfigurationCommand
}

func (st RunConfigurationStateTransitionTestCase) To(state pipelinesv1.SynchronizationState, id string, version string) RunConfigurationStateTransitionTestCase {
	return st.IssuesCommand(SetRunConfigurationStatus{Status: pipelinesv1.Status{
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
		setWorkflowOutput(
			st.SystemStatus.CreateWorkflow(RunConfigurationWorkflowConstants.CreateOperationLabel, phase),
			RunConfigurationWorkflowConstants.RunConfigurationIdParameterName, kfpId),
	)
}

func (st RunConfigurationStateTransitionTestCase) WithUpdateWorkflow(phase argo.WorkflowPhase) RunConfigurationStateTransitionTestCase {
	return st.WithWorkFlow(
		st.SystemStatus.CreateWorkflow(RunConfigurationWorkflowConstants.UpdateOperationLabel, phase),
	)
}

func (st RunConfigurationStateTransitionTestCase) WithUpdateWorkflowWithId(phase argo.WorkflowPhase, kfpId string) RunConfigurationStateTransitionTestCase {
	return st.WithWorkFlow(
		setWorkflowOutput(
			st.SystemStatus.CreateWorkflow(RunConfigurationWorkflowConstants.UpdateOperationLabel, phase),
			RunConfigurationWorkflowConstants.RunConfigurationIdParameterName, kfpId),
	)
}

func (st RunConfigurationStateTransitionTestCase) WithDeletionWorkflow(phase argo.WorkflowPhase) RunConfigurationStateTransitionTestCase {
	return st.WithWorkFlow(
		st.SystemStatus.CreateWorkflow(RunConfigurationWorkflowConstants.DeleteOperationLabel, phase),
	)
}

func (st RunConfigurationStateTransitionTestCase) IssuesCreationWorkflow() RunConfigurationStateTransitionTestCase {
	creationWorkflow := st.workflowFactory.ConstructCreationWorkflow(context.Background(), st.RunConfiguration)
	return st.IssuesCommand(CreateRunConfigurationWorkflow{Workflow: *creationWorkflow})
}

func (st RunConfigurationStateTransitionTestCase) IssuesUpdateWorkflow() RunConfigurationStateTransitionTestCase {
	updateWorkflow := st.workflowFactory.ConstructUpdateWorkflow(context.Background(), st.RunConfiguration)
	return st.IssuesCommand(CreateRunConfigurationWorkflow{Workflow: *updateWorkflow})
}

func (st RunConfigurationStateTransitionTestCase) IssuesDeletionWorkflow() RunConfigurationStateTransitionTestCase {
	deletionWorkflow := st.workflowFactory.ConstructDeletionWorkflow(context.Background(), st.RunConfiguration)
	return st.IssuesCommand(CreateRunConfigurationWorkflow{Workflow: *deletionWorkflow})
}

func (st RunConfigurationStateTransitionTestCase) DeletesAllWorkflows() RunConfigurationStateTransitionTestCase {
	return st.IssuesCommand(DeleteRunConfigurationWorkflows{
		Workflows: st.SystemStatus.Workflows,
	})
}

func (st RunConfigurationStateTransitionTestCase) IssuesCommand(command RunConfigurationCommand) RunConfigurationStateTransitionTestCase {
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
		}
	}

	DescribeTable("State transitions", func(st RunConfigurationStateTransitionTestCase) {
		var stateHandler = RunConfigurationStateHandler{
			WorkflowRepository: st.SystemStatus,
			WorkflowFactory:    workflowFactory,
		}
		commands := stateHandler.StateTransition(context.Background(), st.RunConfiguration)
		is := make([]interface{}, len(st.Commands))
		for i, v := range st.Commands {
			is[i] = v
		}
		Expect(commands).To(ConsistOf(is...))
	},
		Check("Unknown",
			From(pipelinesv1.Unknown, "", "").
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Unknown with version",
			From(pipelinesv1.Unknown, "", v1).
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Unknown with id",
			From(pipelinesv1.Unknown, kfpId, "").
				To(pipelinesv1.Updating, kfpId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Unknown with id and version",
			From(pipelinesv1.Unknown, kfpId, v1).
				To(pipelinesv1.Updating, kfpId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Creating succeeds",
			From(pipelinesv1.Creating, "", v1).
				WithCreateWorkFlowWithId(argo.WorkflowSucceeded, kfpId).
				To(pipelinesv1.Succeeded, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Creating succeeds with existing KfpId",
			From(pipelinesv1.Creating, anotherKfpId, v1).
				WithCreateWorkFlowWithId(argo.WorkflowSucceeded, kfpId).
				To(pipelinesv1.Succeeded, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Creating fails",
			From(pipelinesv1.Creating, "", v1).
				WithCreateWorkFlow(argo.WorkflowFailed).
				To(pipelinesv1.Failed, "", v1).
				DeletesAllWorkflows(),
		),
		Check("Creating without version",
			From(pipelinesv1.Creating, "", "").
				To(pipelinesv1.Failed, "", ""),
		),
		Check("Succeeded no update",
			From(pipelinesv1.Succeeded, kfpId, v1),
		),
		Check("Succeeded with update",
			From(pipelinesv1.Succeeded, kfpId, v0).
				To(pipelinesv1.Updating, kfpId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Succeeded with update but no KfpId",
			From(pipelinesv1.Succeeded, "", v0).
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Succeeded with update but no KfpId and no version",
			From(pipelinesv1.Succeeded, "", "").
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Failed no update",
			From(pipelinesv1.Failed, kfpId, v1),
		),
		Check("Failed with Update",
			From(pipelinesv1.Failed, kfpId, v0).
				To(pipelinesv1.Updating, kfpId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Failed with Update but no KfpId",
			From(pipelinesv1.Failed, "", v0).
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Failed with Update but no KfpId and no version",
			From(pipelinesv1.Failed, "", "").
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Updating succeeds",
			From(pipelinesv1.Updating, anotherKfpId, v1).
				WithUpdateWorkflowWithId(argo.WorkflowSucceeded, kfpId).
				To(pipelinesv1.Succeeded, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Updating fails",
			From(pipelinesv1.Updating, kfpId, v1).
				WithUpdateWorkflow(argo.WorkflowFailed).
				To(pipelinesv1.Failed, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Updating without version",
			From(pipelinesv1.Updating, kfpId, "").
				To(pipelinesv1.Failed, kfpId, ""),
		),
		Check("Updating without KfpId",
			From(pipelinesv1.Updating, "", v1).
				To(pipelinesv1.Failed, "", v1),
		),
		Check("Updating without KfpId or version",
			From(pipelinesv1.Updating, "", "").
				To(pipelinesv1.Failed, "", ""),
		),
		Check("Deleting from Succeeded",
			From(pipelinesv1.Succeeded, kfpId, v1).
				DeletionRequested().
				To(pipelinesv1.Deleting, kfpId, v1).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Failed",
			From(pipelinesv1.Failed, kfpId, v1).
				DeletionRequested().
				To(pipelinesv1.Deleting, kfpId, v1).
				IssuesDeletionWorkflow(),
		),
		Check("Deletion succeeds",
			From(pipelinesv1.Deleting, kfpId, v1).
				DeletionRequested().
				WithDeletionWorkflow(argo.WorkflowSucceeded).
				To(pipelinesv1.Deleted, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Deletion fails",
			From(pipelinesv1.Deleting, kfpId, v1).
				DeletionRequested().
				WithDeletionWorkflow(argo.WorkflowFailed).
				To(pipelinesv1.Deleting, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Stay in deleted",
			From(pipelinesv1.Deleted, kfpId, v1).
				IssuesCommand(DeleteRunConfiguration{}),
		))
})
