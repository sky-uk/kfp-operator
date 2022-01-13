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

type PipelineStateTransitionTestCase struct {
	workflowFactory PipelineWorkflowFactory
	Pipeline        *pipelinesv1.Pipeline
	SystemStatus    StubbedWorkflows
	Commands        []PipelineCommand
}

func (st PipelineStateTransitionTestCase) To(state pipelinesv1.SynchronizationState, id string, version string) PipelineStateTransitionTestCase {
	return st.IssuesCommand(SetPipelineStatus{Status: pipelinesv1.Status{
		KfpId:                id,
		Version:              version,
		SynchronizationState: state,
	}})
}

func (st PipelineStateTransitionTestCase) WithWorkFlow(workflow *argo.Workflow) PipelineStateTransitionTestCase {
	st.SystemStatus.AddWorkflow(*workflow)

	return st
}

func (st PipelineStateTransitionTestCase) WithFailedCreateWorkflow() PipelineStateTransitionTestCase {
	return st.WithWorkFlow(st.SystemStatus.CreateWorkflow(PipelineWorkflowConstants.CreateOperationLabel, argo.WorkflowFailed))
}

func (st PipelineStateTransitionTestCase) WithSucceededCreateWorkflow(kfpId string, version string) PipelineStateTransitionTestCase {
	return st.WithWorkFlow(
		setWorkflowOutputs(
			st.SystemStatus.CreateWorkflow(PipelineWorkflowConstants.CreateOperationLabel, argo.WorkflowSucceeded),
			[]argo.Parameter{
				{
					Name:  PipelineWorkflowConstants.PipelineIdParameterName,
					Value: argo.AnyStringPtr(kfpId),
				},
				{
					Name:  PipelineWorkflowConstants.PipelineVersionParameterName,
					Value: argo.AnyStringPtr(version),
				},
			},
		),
	)
}

func (st PipelineStateTransitionTestCase) WithUpdateWorkflow(phase argo.WorkflowPhase) PipelineStateTransitionTestCase {
	return st.WithWorkFlow(
		st.SystemStatus.CreateWorkflow(PipelineWorkflowConstants.UpdateOperationLabel, phase),
	)
}

func (st PipelineStateTransitionTestCase) WithDeletionWorkflow(phase argo.WorkflowPhase) PipelineStateTransitionTestCase {
	return st.WithWorkFlow(
		st.SystemStatus.CreateWorkflow(PipelineWorkflowConstants.DeleteOperationLabel, phase),
	)
}

func (st PipelineStateTransitionTestCase) IssuesCreationWorkflow() PipelineStateTransitionTestCase {
	creationWorkflow, _ := st.workflowFactory.ConstructCreationWorkflow(context.Background(), st.Pipeline)
	return st.IssuesCommand(CreatePipelineWorkflow{Workflow: *creationWorkflow})
}

func (st PipelineStateTransitionTestCase) IssuesUpdateWorkflow() PipelineStateTransitionTestCase {
	updateWorkflow, _ := st.workflowFactory.ConstructUpdateWorkflow(context.Background(), st.Pipeline)
	return st.IssuesCommand(CreatePipelineWorkflow{Workflow: *updateWorkflow})
}

func (st PipelineStateTransitionTestCase) IssuesDeletionWorkflow() PipelineStateTransitionTestCase {
	deletionWorkflow, _ := st.workflowFactory.ConstructDeletionWorkflow(context.Background(), st.Pipeline)
	return st.IssuesCommand(CreatePipelineWorkflow{Workflow: *deletionWorkflow})
}

func (st PipelineStateTransitionTestCase) DeletesAllWorkflows() PipelineStateTransitionTestCase {
	return st.IssuesCommand(DeletePipelineWorkflows{
		Workflows: st.SystemStatus.Workflows,
	})
}

func (st PipelineStateTransitionTestCase) AcquirePipeline() PipelineStateTransitionTestCase {
	return st.IssuesCommand(AcquirePipeline{})
}

func (st PipelineStateTransitionTestCase) ReleasePipeline() PipelineStateTransitionTestCase {
	return st.IssuesCommand(ReleasePipeline{})
}

func (st PipelineStateTransitionTestCase) IssuesCommand(command PipelineCommand) PipelineStateTransitionTestCase {
	st.Commands = append(st.Commands, command)
	return st
}

func (st PipelineStateTransitionTestCase) DeletionRequested() PipelineStateTransitionTestCase {
	st.Pipeline.DeletionTimestamp = &metav1.Time{time.UnixMilli(1)}
	return st
}

var _ = Describe("Pipeline State handler", func() {

	// TODO: mock workflowFactory
	var workflowFactory = PipelineWorkflowFactory{
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
	specv1 := RandomPipelineSpec()
	v0 := pipelinesv1.RunConfigurationSpec{}.ComputeVersion()
	v1 := specv1.ComputeVersion()
	UnknownState := pipelinesv1.SynchronizationState(RandomString())

	var Check = func(description string, transition PipelineStateTransitionTestCase) TableEntry {
		return Entry(
			description,
			transition,
		)
	}

	var From = func(status pipelinesv1.SynchronizationState, id string, version string) PipelineStateTransitionTestCase {
		pipeline := RandomPipeline()
		pipeline.Spec = specv1
		pipeline.Status = pipelinesv1.Status{
			SynchronizationState: status,
			Version:              version,
			KfpId:                id,
		}

		return PipelineStateTransitionTestCase{
			workflowFactory: workflowFactory,
			Pipeline:        pipeline,
			Commands:        []PipelineCommand{},
		}
	}

	DescribeTable("State transitions", func(st PipelineStateTransitionTestCase) {
		var stateHandler = PipelineStateHandler{
			WorkflowRepository: st.SystemStatus,
			WorkflowFactory:    workflowFactory,
		}
		commands := stateHandler.StateTransition(context.Background(), st.Pipeline)
		Expect(commands).To(Equal(st.Commands))
	},
		Check("Empty",
			From(UnknownState, "", "").
				AcquirePipeline().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Empty with version",
			From(UnknownState, "", v1).
				AcquirePipeline().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Empty with id",
			From(UnknownState, kfpId, "").
				AcquirePipeline().
				To(pipelinesv1.Updating, kfpId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Empty with id and version",
			From(UnknownState, kfpId, v1).
				AcquirePipeline().
				To(pipelinesv1.Updating, kfpId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Creating succeeds with kfpId and version",
			From(pipelinesv1.Creating, "", v1).
				AcquirePipeline().
				WithSucceededCreateWorkflow(kfpId, v1).
				To(pipelinesv1.Succeeded, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Creating succeeds with different KfpId and version",
			From(pipelinesv1.Creating, anotherKfpId, v1).
				AcquirePipeline().
				WithSucceededCreateWorkflow(kfpId, v1).
				To(pipelinesv1.Succeeded, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Creating succeeds with KfpId but no version",
			From(pipelinesv1.Creating, "", v1).
				AcquirePipeline().
				WithSucceededCreateWorkflow(kfpId, "").
				To(pipelinesv1.Failed, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Creating fails",
			From(pipelinesv1.Creating, "", v1).
				AcquirePipeline().
				WithFailedCreateWorkflow().
				To(pipelinesv1.Failed, "", v1).
				DeletesAllWorkflows(),
		),
		Check("Creating without version",
			From(pipelinesv1.Creating, "", "").
				AcquirePipeline().
				To(pipelinesv1.Failed, "", ""),
		),
		Check("Succeeded no update",
			From(pipelinesv1.Succeeded, kfpId, v1).
				AcquirePipeline(),
		),
		Check("Succeeded with update",
			From(pipelinesv1.Succeeded, kfpId, v0).
				AcquirePipeline().
				To(pipelinesv1.Updating, kfpId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Succeeded with update but no KfpId",
			From(pipelinesv1.Succeeded, "", v0).
				AcquirePipeline().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Succeeded with update but no KfpId and no version",
			From(pipelinesv1.Succeeded, "", "").
				AcquirePipeline().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Failed no update",
			From(pipelinesv1.Failed, kfpId, v1).
				AcquirePipeline(),
		),
		Check("Failed with Update",
			From(pipelinesv1.Failed, kfpId, v0).
				AcquirePipeline().
				To(pipelinesv1.Updating, kfpId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Failed with Update but no KfpId",
			From(pipelinesv1.Failed, "", v0).
				AcquirePipeline().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Failed with Update but no KfpId and no version",
			From(pipelinesv1.Failed, "", "").
				AcquirePipeline().
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Updating succeeds",
			From(pipelinesv1.Updating, kfpId, v1).
				AcquirePipeline().
				WithUpdateWorkflow(argo.WorkflowSucceeded).
				To(pipelinesv1.Succeeded, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Updating fails",
			From(pipelinesv1.Updating, kfpId, v1).
				AcquirePipeline().
				WithUpdateWorkflow(argo.WorkflowFailed).
				To(pipelinesv1.Failed, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Updating without version",
			From(pipelinesv1.Updating, kfpId, "").
				AcquirePipeline().
				To(pipelinesv1.Failed, kfpId, ""),
		),
		Check("Updating without KfpId",
			From(pipelinesv1.Updating, "", v1).
				AcquirePipeline().
				To(pipelinesv1.Failed, "", v1),
		),
		Check("Updating without KfpId or version",
			From(pipelinesv1.Updating, "", "").
				AcquirePipeline().
				To(pipelinesv1.Failed, "", ""),
		),
		Check("Deleting from Succeeded",
			From(pipelinesv1.Succeeded, kfpId, v1).
				AcquirePipeline().
				DeletionRequested().
				To(pipelinesv1.Deleting, kfpId, v1).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Succeeded without kfpId",
			From(pipelinesv1.Succeeded, "", v1).
				AcquirePipeline().
				DeletionRequested().
				To(pipelinesv1.Deleted, "", v1),
		),
		Check("Deleting from Failed",
			From(pipelinesv1.Failed, kfpId, v1).
				AcquirePipeline().
				DeletionRequested().
				To(pipelinesv1.Deleting, kfpId, v1).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Failed without kfpId",
			From(pipelinesv1.Failed, "", v1).
				AcquirePipeline().
				DeletionRequested().
				To(pipelinesv1.Deleted, "", v1),
		),
		Check("Deletion succeeds",
			From(pipelinesv1.Deleting, kfpId, v1).
				AcquirePipeline().
				DeletionRequested().
				WithDeletionWorkflow(argo.WorkflowSucceeded).
				To(pipelinesv1.Deleted, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Deletion fails",
			From(pipelinesv1.Deleting, kfpId, v1).
				AcquirePipeline().
				DeletionRequested().
				WithDeletionWorkflow(argo.WorkflowFailed).
				To(pipelinesv1.Deleting, kfpId, v1).
				DeletesAllWorkflows(),
		),
		Check("Stay in deleted",
			From(pipelinesv1.Deleted, kfpId, v1).
				ReleasePipeline(),
		),
	)
})
