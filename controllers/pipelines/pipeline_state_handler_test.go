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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type PipelineStateTransitionTestCase struct {
	workflowFactory WorkflowFactory[*pipelinesv1.Pipeline]
	Pipeline        *pipelinesv1.Pipeline
	SystemStatus    StubbedWorkflows
	Commands        []Command
}

type FailingPipelineWorkflowFactory struct{}

func (f FailingPipelineWorkflowFactory) ConstructCreationWorkflow(_ *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	return nil, fmt.Errorf("an error occurred")
}

func (f FailingPipelineWorkflowFactory) ConstructUpdateWorkflow(_ *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	return nil, fmt.Errorf("an error occurred")
}

func (f FailingPipelineWorkflowFactory) ConstructDeletionWorkflow(_ *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	return nil, fmt.Errorf("an error occurred")
}

func (st PipelineStateTransitionTestCase) WorkflowConstructionFails() PipelineStateTransitionTestCase {
	st.workflowFactory = FailingPipelineWorkflowFactory{}
	return st
}

func (st PipelineStateTransitionTestCase) WithWorkFlow(workflow *argo.Workflow) PipelineStateTransitionTestCase {
	st.SystemStatus.AddWorkflow(*workflow)

	return st
}

func (st PipelineStateTransitionTestCase) WithFailedCreateWorkflow() PipelineStateTransitionTestCase {
	return st.WithWorkFlow(CreateTestWorkflow(WorkflowConstants.CreateOperationLabel, argo.WorkflowFailed))
}

func (st PipelineStateTransitionTestCase) WithSucceededCreateWorkflow(kfpId string, version string) PipelineStateTransitionTestCase {
	return st.WithWorkFlow(
		setWorkflowOutputs(
			CreateTestWorkflow(WorkflowConstants.CreateOperationLabel, argo.WorkflowSucceeded),
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
		CreateTestWorkflow(WorkflowConstants.UpdateOperationLabel, phase),
	)
}

func (st PipelineStateTransitionTestCase) WithDeletionWorkflow(phase argo.WorkflowPhase) PipelineStateTransitionTestCase {
	return st.WithWorkFlow(
		CreateTestWorkflow(WorkflowConstants.DeleteOperationLabel, phase),
	)
}

func (st PipelineStateTransitionTestCase) IssuesCreationWorkflow() PipelineStateTransitionTestCase {
	creationWorkflow, _ := st.workflowFactory.ConstructCreationWorkflow(st.Pipeline)
	return st.IssuesCommand(CreateWorkflow{Workflow: *creationWorkflow})
}

func (st PipelineStateTransitionTestCase) IssuesUpdateWorkflow() PipelineStateTransitionTestCase {
	updateWorkflow, _ := st.workflowFactory.ConstructUpdateWorkflow(st.Pipeline)
	return st.IssuesCommand(CreateWorkflow{Workflow: *updateWorkflow})
}

func (st PipelineStateTransitionTestCase) IssuesDeletionWorkflow() PipelineStateTransitionTestCase {
	deletionWorkflow, _ := st.workflowFactory.ConstructDeletionWorkflow(st.Pipeline)
	return st.IssuesCommand(CreateWorkflow{Workflow: *deletionWorkflow})
}

func (st PipelineStateTransitionTestCase) DeletesAllWorkflows() PipelineStateTransitionTestCase {
	return st.IssuesCommand(DeleteWorkflows{
		Workflows: st.SystemStatus.Workflows,
	})
}

func (st PipelineStateTransitionTestCase) AcquirePipeline() PipelineStateTransitionTestCase {
	return st.IssuesCommand(AcquireResource{})
}

func (st PipelineStateTransitionTestCase) ReleasePipeline() PipelineStateTransitionTestCase {
	return st.IssuesCommand(ReleaseResource{})
}

func (st PipelineStateTransitionTestCase) IssuesCommand(command Command) PipelineStateTransitionTestCase {
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
		WorkflowFactoryBase: WorkflowFactoryBase{
			Config: config.Configuration{
				KfpEndpoint: "http://www.example.com",
			},
		},
	}

	kfpId := "12345"
	anotherKfpId := "67890"
	specv1 := pipelinesv1.RandomPipelineSpec()
	v0 := pipelinesv1.PipelineSpec{}.ComputeVersion()
	v1 := specv1.ComputeVersion()
	UnknownState := apis.SynchronizationState(apis.RandomString())

	var Check = func(description string, transition PipelineStateTransitionTestCase) TableEntry {
		return Entry(
			description,
			transition,
		)
	}

	var From = func(status apis.SynchronizationState, id string, version string) PipelineStateTransitionTestCase {
		pipeline := pipelinesv1.RandomPipeline()
		pipeline.Spec = specv1
		pipeline.Status = apis.Status{
			SynchronizationState: status,
			Version:              version,
			KfpId:                id,
		}

		return PipelineStateTransitionTestCase{
			workflowFactory: workflowFactory,
			Pipeline:        pipeline,
			Commands:        []Command{},
		}
	}

	DescribeTable("State transitions", func(st PipelineStateTransitionTestCase) {
		var stateHandler = PipelineStateHandler{
			WorkflowRepository: st.SystemStatus,
			WorkflowFactory:    st.workflowFactory,
		}
		commands := stateHandler.StateTransition(context.Background(), st.Pipeline)
		Expect(commands).To(Equal(st.Commands))
	},
		Check("Empty",
			From(UnknownState, "", "").
				AcquirePipeline().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Empty and workflow creation fails",
			From(UnknownState, "", "").
				AcquirePipeline().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Empty with version",
			From(UnknownState, "", v1).
				AcquirePipeline().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Empty with id",
			From(UnknownState, kfpId, "").
				AcquirePipeline().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Empty with id and workflow creation fails",
			From(UnknownState, kfpId, "").
				AcquirePipeline().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Empty with id and version",
			From(UnknownState, kfpId, v1).
				AcquirePipeline().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Creating succeeds with kfpId and version",
			From(apis.Creating, "", v1).
				AcquirePipeline().
				WithSucceededCreateWorkflow(kfpId, v1).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Succeeded).
					WithKfpId(kfpId).
					WithVersion(v1)).
				DeletesAllWorkflows(),
		),
		Check("Creating succeeds with different KfpId and version",
			From(apis.Creating, anotherKfpId, v1).
				AcquirePipeline().
				WithSucceededCreateWorkflow(kfpId, v1).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Succeeded).
					WithKfpId(kfpId).
					WithVersion(v1)).
				DeletesAllWorkflows(),
		),
		Check("Creating succeeds with KfpId but no version",
			From(apis.Creating, "", v1).
				AcquirePipeline().
				WithSucceededCreateWorkflow(kfpId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithKfpId(kfpId).WithVersion(v1).
					WithMessage("pipeline creation succeeded but version upload failed")).
				DeletesAllWorkflows(),
		),
		Check("Creating fails",
			From(apis.Creating, "", v1).
				AcquirePipeline().
				WithFailedCreateWorkflow().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithVersion(v1).
					WithMessage("pipeline creation failed")).
				DeletesAllWorkflows(),
		),
		Check("Creating without version",
			From(apis.Creating, "", "").
				AcquirePipeline().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithMessage("creating pipeline with empty version")),
		),
		Check("Succeeded no update",
			From(apis.Succeeded, kfpId, v1).
				AcquirePipeline(),
		),
		Check("Succeeded with update",
			From(apis.Succeeded, kfpId, v0).
				AcquirePipeline().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Succeeded with update and workflow creation fails",
			From(apis.Succeeded, kfpId, v0).
				AcquirePipeline().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Succeeded with update but no KfpId",
			From(apis.Succeeded, "", v0).
				AcquirePipeline().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Succeeded with update but no kfpId and workflow creation fails",
			From(apis.Succeeded, "", v0).
				AcquirePipeline().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Succeeded with update but no KfpId and no version",
			From(apis.Succeeded, "", "").
				AcquirePipeline().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Failed no update",
			From(apis.Failed, kfpId, v1).
				AcquirePipeline(),
		),
		Check("Failed with update",
			From(apis.Failed, kfpId, v0).
				AcquirePipeline().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Failed with update and workflow creation fails",
			From(apis.Failed, kfpId, v0).
				AcquirePipeline().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Failed with update but no KfpId",
			From(apis.Failed, "", v0).
				AcquirePipeline().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Failed with update but no KfpId and workflow creation fails",
			From(apis.Failed, "", v0).
				AcquirePipeline().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Failed with update but no KfpId and no version",
			From(apis.Failed, "", "").
				AcquirePipeline().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Updating succeeds",
			From(apis.Updating, kfpId, v1).
				AcquirePipeline().
				WithUpdateWorkflow(argo.WorkflowSucceeded).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Succeeded).
					WithKfpId(kfpId).
					WithVersion(v1)).
				DeletesAllWorkflows(),
		),
		Check("Updating fails",
			From(apis.Updating, kfpId, v1).
				AcquirePipeline().
				WithUpdateWorkflow(argo.WorkflowFailed).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage("pipeline update failed")).
				DeletesAllWorkflows(),
		),
		Check("Updating without version",
			From(apis.Updating, kfpId, "").
				AcquirePipeline().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithKfpId(kfpId).
					WithMessage("updating pipeline with empty version or kfpId")),
		),
		Check("Updating without KfpId",
			From(apis.Updating, "", v1).
				AcquirePipeline().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithVersion(v1).
					WithMessage("updating pipeline with empty version or kfpId")),
		),
		Check("Updating without KfpId or version",
			From(apis.Updating, "", "").
				AcquirePipeline().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithMessage("updating pipeline with empty version or kfpId")),
		),
		Check("Deleting from Succeeded",
			From(apis.Succeeded, kfpId, v1).
				AcquirePipeline().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Succeeded without kfpId",
			From(apis.Succeeded, "", v1).
				AcquirePipeline().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleted).
					WithVersion(v1)),
		),
		Check("Deleting from Failed",
			From(apis.Failed, kfpId, v1).
				AcquirePipeline().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Failed without kfpId",
			From(apis.Failed, "", v1).
				AcquirePipeline().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleted).
					WithVersion(v1)),
		),
		Check("Deletion succeeds",
			From(apis.Deleting, kfpId, v1).
				AcquirePipeline().
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
				AcquirePipeline().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage("pipeline deletion failed")).
				WithDeletionWorkflow(argo.WorkflowFailed).
				DeletesAllWorkflows(),
		),
		Check("Stay in deleted",
			From(apis.Deleted, kfpId, v1).
				ReleasePipeline(),
		),
	)
})
