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

type ExperimentStateTransitionTestCase struct {
	workflowFactory ExperimentWorkflowFactory
	Experiment      *pipelinesv1.Experiment
	SystemStatus    StubbedWorkflows
	Commands        []ExperimentCommand
}

func (st ExperimentStateTransitionTestCase) To(state pipelinesv1.SynchronizationState, id string, version string) ExperimentStateTransitionTestCase {
	return st.IssuesCommand(SetExperimentStatus{Status: pipelinesv1.Status{
		KfpId:                id,
		Version:              version,
		SynchronizationState: state,
	}})
}

func (st ExperimentStateTransitionTestCase) WithWorkFlow(workflow *argo.Workflow) ExperimentStateTransitionTestCase {
	st.SystemStatus.AddWorkflow(*workflow)
	return st
}

func (st ExperimentStateTransitionTestCase) WithCreateWorkFlow(phase argo.WorkflowPhase) ExperimentStateTransitionTestCase {
	return st.WithWorkFlow(st.SystemStatus.CreateWorkflow(ExperimentWorkflowConstants.CreateOperationLabel, phase))
}

func (st ExperimentStateTransitionTestCase) WithCreateWorkFlowWithId(phase argo.WorkflowPhase, kfpId string) ExperimentStateTransitionTestCase {
	return st.WithWorkFlow(
		setWorkflowOutputs(
			st.SystemStatus.CreateWorkflow(ExperimentWorkflowConstants.CreateOperationLabel, phase),
			[]argo.Parameter{
				{
					Name:  ExperimentWorkflowConstants.ExperimentIdParameterName,
					Value: argo.AnyStringPtr(kfpId),
				},
			},
		),
	)
}

func (st ExperimentStateTransitionTestCase) WithFailedUpdateWorkflow() ExperimentStateTransitionTestCase {
	return st.WithWorkFlow(
		st.SystemStatus.CreateWorkflow(ExperimentWorkflowConstants.UpdateOperationLabel, argo.WorkflowFailed),
	)
}

func (st ExperimentStateTransitionTestCase) WithSucceededUpdateWorkflowWithId(kfpId string) ExperimentStateTransitionTestCase {
	return st.WithWorkFlow(
		setWorkflowOutputs(
			st.SystemStatus.CreateWorkflow(ExperimentWorkflowConstants.UpdateOperationLabel, argo.WorkflowSucceeded),
			[]argo.Parameter{
				{
					Name:  ExperimentWorkflowConstants.ExperimentIdParameterName,
					Value: argo.AnyStringPtr(kfpId),
				},
			},
		),
	)
}

func (st ExperimentStateTransitionTestCase) WithDeletionWorkflow(phase argo.WorkflowPhase) ExperimentStateTransitionTestCase {
	return st.WithWorkFlow(
		st.SystemStatus.CreateWorkflow(ExperimentWorkflowConstants.DeleteOperationLabel, phase),
	)
}

func (st ExperimentStateTransitionTestCase) IssuesCreationWorkflow() ExperimentStateTransitionTestCase {
	creationWorkflow, _ := st.workflowFactory.ConstructCreationWorkflow(context.Background(), st.Experiment)
	return st.IssuesCommand(CreateExperimentWorkflow{Workflow: *creationWorkflow})
}

func (st ExperimentStateTransitionTestCase) IssuesUpdateWorkflow() ExperimentStateTransitionTestCase {
	updateWorkflow, _ := st.workflowFactory.ConstructUpdateWorkflow(context.Background(), st.Experiment)
	return st.IssuesCommand(CreateExperimentWorkflow{Workflow: *updateWorkflow})
}

func (st ExperimentStateTransitionTestCase) IssuesDeletionWorkflow() ExperimentStateTransitionTestCase {
	deletionWorkflow, _ := st.workflowFactory.ConstructDeletionWorkflow(context.Background(), st.Experiment)
	return st.IssuesCommand(CreateExperimentWorkflow{Workflow: *deletionWorkflow})
}

func (st ExperimentStateTransitionTestCase) DeletesAllWorkflows() ExperimentStateTransitionTestCase {
	return st.IssuesCommand(DeleteExperimentWorkflows{
		Workflows: st.SystemStatus.Workflows,
	})
}

func (st ExperimentStateTransitionTestCase) AcquireExperiment() ExperimentStateTransitionTestCase {
	return st.IssuesCommand(AcquireExperiment{})
}

func (st ExperimentStateTransitionTestCase) ReleaseExperiment() ExperimentStateTransitionTestCase {
	return st.IssuesCommand(ReleaseExperiment{})
}

func (st ExperimentStateTransitionTestCase) IssuesCommand(command ExperimentCommand) ExperimentStateTransitionTestCase {
	st.Commands = append(st.Commands, command)
	return st
}

func (st ExperimentStateTransitionTestCase) DeletionRequested() ExperimentStateTransitionTestCase {
	st.Experiment.DeletionTimestamp = &metav1.Time{time.UnixMilli(1)}
	return st
}

var _ = Describe("Experiment State handler", func() {
	// TODO: mock workflowFactory
	var workflowFactory = ExperimentWorkflowFactory{
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
	specv1 := RandomExperimentSpec()
	v0 := pipelinesv1.ExperimentSpec{}.ComputeVersion()
	v1 := specv1.ComputeVersion()
	UnknownState := pipelinesv1.SynchronizationState(RandomString())

	var Check = func(description string, transition ExperimentStateTransitionTestCase) TableEntry {
		return Entry(
			description,
			transition,
		)
	}

	var From = func(status pipelinesv1.SynchronizationState, id string, version string) ExperimentStateTransitionTestCase {
		experiment := RandomExperiment()
		experiment.Spec = specv1
		experiment.Status = pipelinesv1.Status{
			SynchronizationState: status,
			Version:              version,
			KfpId:                id,
		}

		return ExperimentStateTransitionTestCase{
			workflowFactory: workflowFactory,
			Experiment:      experiment,
			Commands:        []ExperimentCommand{},
		}
	}

	DescribeTable("State transitions", func(st ExperimentStateTransitionTestCase) {
		var stateHandler = ExperimentStateHandler{
			WorkflowRepository: st.SystemStatus,
			WorkflowFactory:    workflowFactory,
		}
		commands := stateHandler.StateTransition(context.Background(), st.Experiment)
		Expect(commands).To(Equal(st.Commands))
	},
		Check("Empty",
			From(UnknownState, "", "").
				AcquireExperiment().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Empty with version",
			From(UnknownState, "", v1).
				AcquireExperiment().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Empty with id",
			From(UnknownState, kfpId, "").
				AcquireExperiment().
				To(pipelinesv1.Updating, kfpId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Empty with id and version",
			From(UnknownState, kfpId, v1).
				AcquireExperiment().
				To(pipelinesv1.Updating, kfpId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Creating succeeds",
			From(pipelinesv1.Creating, "", v1).
				AcquireExperiment().
				WithCreateWorkFlowWithId(argo.WorkflowSucceeded, kfpId).
				To(pipelinesv1.Succeeded, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Creating succeeds with existing KfpId",
			From(pipelinesv1.Creating, anotherKfpId, v1).
				AcquireExperiment().
				WithCreateWorkFlowWithId(argo.WorkflowSucceeded, kfpId).
				To(pipelinesv1.Succeeded, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Creating fails",
			From(pipelinesv1.Creating, "", v1).
				AcquireExperiment().
				WithCreateWorkFlow(argo.WorkflowFailed).
				To(pipelinesv1.Failed, "", v1).
				DeletesAllWorkflows(),
		),
		Check("Creating without version",
			From(pipelinesv1.Creating, "", "").
				AcquireExperiment().
				To(pipelinesv1.Failed, "", ""),
		),
		Check("Succeeded no update",
			From(pipelinesv1.Succeeded, kfpId, v1).
				AcquireExperiment(),
		),
		Check("Succeeded with update",
			From(pipelinesv1.Succeeded, kfpId, v0).
				AcquireExperiment().
				To(pipelinesv1.Updating, kfpId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Succeeded with update but no KfpId",
			From(pipelinesv1.Succeeded, "", v0).
				AcquireExperiment().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Succeeded with update but no KfpId and no version",
			From(pipelinesv1.Succeeded, "", "").
				AcquireExperiment().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Failed no update",
			From(pipelinesv1.Failed, kfpId, v1).
				AcquireExperiment(),
		),
		Check("Failed with Update",
			From(pipelinesv1.Failed, kfpId, v0).
				AcquireExperiment().
				To(pipelinesv1.Updating, kfpId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Failed with Update but no KfpId",
			From(pipelinesv1.Failed, "", v0).
				AcquireExperiment().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Failed with Update but no KfpId and no version",
			From(pipelinesv1.Failed, "", "").
				AcquireExperiment().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Updating succeeds with kfpId",
			From(pipelinesv1.Updating, anotherKfpId, v1).
				AcquireExperiment().
				WithSucceededUpdateWorkflowWithId(kfpId).
				To(pipelinesv1.Succeeded, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Updating succeeds without kfpId",
			From(pipelinesv1.Updating, anotherKfpId, v1).
				AcquireExperiment().
				WithSucceededUpdateWorkflowWithId("").
				To(pipelinesv1.Failed, "", v1).
				DeletesAllWorkflows(),
		),
		Check("Updating fails",
			From(pipelinesv1.Updating, kfpId, v1).
				AcquireExperiment().
				WithFailedUpdateWorkflow().
				To(pipelinesv1.Failed, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Updating without version",
			From(pipelinesv1.Updating, kfpId, "").
				AcquireExperiment().
				To(pipelinesv1.Failed, kfpId, ""),
		),
		Check("Updating without KfpId",
			From(pipelinesv1.Updating, "", v1).
				AcquireExperiment().
				To(pipelinesv1.Failed, "", v1),
		),
		Check("Updating without KfpId or version",
			From(pipelinesv1.Updating, "", "").
				AcquireExperiment().
				To(pipelinesv1.Failed, "", ""),
		),
		Check("Deleting from Succeeded",
			From(pipelinesv1.Succeeded, kfpId, v1).
				AcquireExperiment().
				DeletionRequested().
				To(pipelinesv1.Deleting, kfpId, v1).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Succeeded without kfpId",
			From(pipelinesv1.Succeeded, "", v1).
				AcquireExperiment().
				DeletionRequested().
				To(pipelinesv1.Deleted, "", v1),
		),
		Check("Deleting from Failed",
			From(pipelinesv1.Failed, kfpId, v1).
				AcquireExperiment().
				DeletionRequested().
				To(pipelinesv1.Deleting, kfpId, v1).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Failed without kfpId",
			From(pipelinesv1.Failed, "", v1).
				AcquireExperiment().
				DeletionRequested().
				To(pipelinesv1.Deleted, "", v1),
		),
		Check("Deletion succeeds",
			From(pipelinesv1.Deleting, kfpId, v1).
				AcquireExperiment().
				DeletionRequested().
				WithDeletionWorkflow(argo.WorkflowSucceeded).
				To(pipelinesv1.Deleted, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Deletion fails",
			From(pipelinesv1.Deleting, kfpId, v1).
				AcquireExperiment().
				DeletionRequested().
				WithDeletionWorkflow(argo.WorkflowFailed).
				To(pipelinesv1.Deleting, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Stay in deleted",
			From(pipelinesv1.Deleted, kfpId, v1).
				ReleaseExperiment(),
		))
})