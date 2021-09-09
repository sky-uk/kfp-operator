package controllers

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// +kubebuilder:scaffold:imports
)

var now = metav1.Now()
var workflows = Workflows{
	Config: WorkflowConfiguration{
		CompilerImage: "image:v1",
		UploaderImage: "image:v1",
	},
}
var stateHandler = StateHandler{
	Workflows: workflows,
}

type StubbedWorkflows struct {
	Workflows []argo.Workflow
}

func (sw StubbedWorkflows) GetByOperation(operation string) []argo.Workflow {
	return sw.Workflows
}

type TestCase struct {
	Pipeline  *pipelinesv1.Pipeline
	Workflows []argo.Workflow
	Commands  []Command
}

func From(status pipelinesv1.SynchronizationState, id string, version string) TestCase {
	pipeline := randomPipeline()
	pipeline.Status = pipelinesv1.PipelineStatus{
		SynchronizationState: status,
		Version:              version,
		Id:                   id,
	}

	return TestCase{
		Pipeline: pipeline,
	}
}

func (st TestCase) To(state pipelinesv1.SynchronizationState, id string, version string) TestCase {
	return st.IssuesCommand(SetPipelineStatus{Status: pipelinesv1.PipelineStatus{
		Id:                   id,
		Version:              version,
		SynchronizationState: state,
	}})
}

func (st TestCase) DeletionRequested() TestCase {
	st.Pipeline.DeletionTimestamp = &now
	return st
}

func createWorkflow(operation string, phase argo.WorkflowPhase) *argo.Workflow {
	return &argo.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operation + "-pipeline",
			Namespace: "default",
			Labels: map[string]string{
				OperationLabelKey: operation,
			},
		},
		Status: argo.WorkflowStatus{
			Phase: phase,
		},
	}
}

func (st TestCase) WithWorkFlow(workflow *argo.Workflow) TestCase {
	st.Workflows = append(st.Workflows, *workflow)
	return st
}

func (st TestCase) IssuesCreationWorkflow() TestCase {
	creationWorkflow, _ := workflows.ConstructCreationWorkflow(st.Pipeline)
	return st.IssuesCommand(CreateWorkflow{Workflow: *creationWorkflow})
}

func (st TestCase) IssuesUpdateWorkflow() TestCase {
	updateWorkflow, _ := workflows.ConstructUpdateWorkflow(st.Pipeline)
	return st.IssuesCommand(CreateWorkflow{Workflow: *updateWorkflow})
}

func (st TestCase) IssuesDeletionWorkflow() TestCase {
	deletionWorkflow := workflows.ConstructDeletionWorkflow(st.Pipeline)
	return st.IssuesCommand(CreateWorkflow{Workflow: *deletionWorkflow})
}

func (st TestCase) DeletesAllWorkflows() TestCase {
	return st.IssuesCommand(DeleteWorkflows{
		Workflows: st.Workflows,
	})
}

func (st TestCase) IssuesCommand(commdand Command) TestCase {
	st.Commands = append(st.Commands, commdand)
	return st
}

func Check(description string, transition TestCase) TableEntry {
	return Entry(
		description,
		transition,
	)
}

var _ = Describe("Pipeline State handler", func() {

	DescribeTable("State transitions", func(st TestCase) {
		commands := stateHandler.StateTransition(st.Pipeline, StubbedWorkflows{st.Workflows})
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
			From(pipelinesv1.Unknown, PipelineId, "").
				To(pipelinesv1.Updating, PipelineId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Unknown with id and version",
			From(pipelinesv1.Unknown, PipelineId, v1).
				To(pipelinesv1.Updating, PipelineId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Creation succeeds",
			From(pipelinesv1.Creating, "", v1).
				WithWorkFlow(setWorkflowOutput(createWorkflow(Create, argo.WorkflowSucceeded), PipelineIdParameterName, PipelineId)).
				To(pipelinesv1.Succeeded, PipelineId, v1).
				DeletesAllWorkflows(),
		),
		Check("Creation succeeds with existing Id",
			From(pipelinesv1.Creating, AnotherPipelineId, v1).
				WithWorkFlow(setWorkflowOutput(createWorkflow(Create, argo.WorkflowSucceeded), PipelineIdParameterName, PipelineId)).
				To(pipelinesv1.Succeeded, PipelineId, v1).
				DeletesAllWorkflows(),
		),
		Check("Creation fails with Id",
			From(pipelinesv1.Creating, "", v1).
				WithWorkFlow(setWorkflowOutput(createWorkflow(Create, argo.WorkflowFailed), PipelineIdParameterName, PipelineId)).
				To(pipelinesv1.Failed, PipelineId, v1).
				DeletesAllWorkflows(),
		),
		Check("Creation fails",
			From(pipelinesv1.Creating, "", v1).
				WithWorkFlow(createWorkflow(Create, argo.WorkflowFailed)).
				To(pipelinesv1.Failed, "", v1).
				DeletesAllWorkflows(),
		),
		Check("Creating without version",
			From(pipelinesv1.Creating, "", "").
				To(pipelinesv1.Failed, "", ""),
		),
		Check("Succeeded no update",
			From(pipelinesv1.Succeeded, PipelineId, v1),
		),
		Check("Succeeded with update",
			From(pipelinesv1.Succeeded, PipelineId, v0).
				To(pipelinesv1.Updating, PipelineId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Succeeded with update but no Id",
			From(pipelinesv1.Succeeded, "", v0).
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Succeeded with update but no Id and no version",
			From(pipelinesv1.Succeeded, "", "").
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Failed no update",
			From(pipelinesv1.Failed, PipelineId, v1),
		),
		Check("Failed with Update",
			From(pipelinesv1.Failed, PipelineId, v0).
				To(pipelinesv1.Updating, PipelineId, v1).
				IssuesUpdateWorkflow(),
		),
		Check("Failed with Update but no Id",
			From(pipelinesv1.Failed, "", v0).
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Failed with Update but no Id and no version",
			From(pipelinesv1.Failed, "", "").
				To(pipelinesv1.Creating, "", v1).
				IssuesCreationWorkflow(),
		),
		Check("Updating succeeds",
			From(pipelinesv1.Updating, PipelineId, v1).
				WithWorkFlow(createWorkflow(Update, argo.WorkflowSucceeded)).
				To(pipelinesv1.Succeeded, PipelineId, v1).
				DeletesAllWorkflows(),
		),
		Check("Updating fails",
			From(pipelinesv1.Updating, PipelineId, v1).
				WithWorkFlow(createWorkflow(Update, argo.WorkflowFailed)).
				To(pipelinesv1.Failed, PipelineId, v1).
				DeletesAllWorkflows(),
		),
		Check("updating without version",
			From(pipelinesv1.Updating, PipelineId, "").
				To(pipelinesv1.Failed, PipelineId, ""),
		),
		Check("updating without version",
			From(pipelinesv1.Updating, PipelineId, "").
				To(pipelinesv1.Failed, PipelineId, ""),
		),
		Check("updating without Id",
			From(pipelinesv1.Updating, "", v1).
				To(pipelinesv1.Failed, "", v1),
		),
		Check("updating without Id or version",
			From(pipelinesv1.Updating, "", "").
				To(pipelinesv1.Failed, "", ""),
		),
		Check("Deleting from Succeeded",
			From(pipelinesv1.Succeeded, PipelineId, v1).
				DeletionRequested().
				To(pipelinesv1.Deleting, PipelineId, v1).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Failed",
			From(pipelinesv1.Failed, PipelineId, v1).
				DeletionRequested().
				To(pipelinesv1.Deleting, PipelineId, v1).
				IssuesDeletionWorkflow(),
		),
		Check("Deletion succeeds",
			From(pipelinesv1.Deleting, PipelineId, v1).
				DeletionRequested().
				WithWorkFlow(createWorkflow(Delete, argo.WorkflowSucceeded)).
				To(pipelinesv1.Deleted, PipelineId, v1).
				DeletesAllWorkflows(),
		),
		Check("Deletion fails",
			From(pipelinesv1.Deleting, PipelineId, v1).
				DeletionRequested().
				WithWorkFlow(createWorkflow(Delete, argo.WorkflowFailed)).
				To(pipelinesv1.Deleting, PipelineId, v1).
				DeletesAllWorkflows(),
		),
		Check("Stay in deleted",
			From(pipelinesv1.Deleted, PipelineId, v1).
				IssuesCommand(DeletePipeline{}),
		),
	)
})
