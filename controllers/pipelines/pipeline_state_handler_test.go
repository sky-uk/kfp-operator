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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// +kubebuilder:scaffold:imports
)

var now = metav1.Now()

// TODO: mock workflowFactory
var workflowFactory = WorkflowFactory{
	Config: configv1.Configuration{
		KfpToolsImage:   "kfp-tools",
		CompilerImage:   "compiler",
		ImagePullPolicy: "Never",
		KfpEndpoint:     "http://www.example.com",
	},
}

type StubbedWorkflows struct {
	Workflows []argo.Workflow
}

func (sw StubbedWorkflows) GetByOperation(ctx context.Context, operation string, pipeline *pipelinesv1.Pipeline) []argo.Workflow {
	return sw.Workflows
}

type TestCase struct {
	Pipeline  *pipelinesv1.Pipeline
	Workflows []argo.Workflow
	Commands  []Command
}

func From(status pipelinesv1.SynchronizationState, id string, version string) TestCase {
	pipeline := RandomPipeline()
	pipeline.Spec = SpecV1
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

func (st TestCase) IssuesCreationWorkflow(version string) TestCase {
	creationWorkflow, _ := workflowFactory.ConstructCreationWorkflow(st.Pipeline.Spec, st.Pipeline.ObjectMeta, version)
	return st.IssuesCommand(CreateWorkflow{Workflow: *creationWorkflow})
}

func (st TestCase) IssuesUpdateWorkflow(version string) TestCase {
	updateWorkflow, _ := workflowFactory.ConstructUpdateWorkflow(st.Pipeline.Spec, st.Pipeline.ObjectMeta, st.Pipeline.Status.Id, version)
	return st.IssuesCommand(CreateWorkflow{Workflow: *updateWorkflow})
}

func (st TestCase) IssuesDeletionWorkflow() TestCase {
	deletionWorkflow := workflowFactory.ConstructDeletionWorkflow(st.Pipeline.ObjectMeta, st.Pipeline.Status.Id)
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
		var stateHandler = StateHandler{
			WorkflowRepository: StubbedWorkflows{st.Workflows},
			WorkflowFactory:    workflowFactory,
		}
		commands := stateHandler.StateTransition(context.Background(), st.Pipeline)
		is := make([]interface{}, len(st.Commands))
		for i, v := range st.Commands {
			is[i] = v
		}
		Expect(commands).To(ConsistOf(is...))
	},
		Check("Unknown",
			From(pipelinesv1.Unknown, "", "").
				To(pipelinesv1.Creating, "", V1).
				IssuesCreationWorkflow(V1),
		),
		Check("Unknown with version",
			From(pipelinesv1.Unknown, "", V1).
				To(pipelinesv1.Creating, "", V1).
				IssuesCreationWorkflow(V1),
		),
		Check("Unknown with id",
			From(pipelinesv1.Unknown, PipelineId, "").
				To(pipelinesv1.Updating, PipelineId, V1).
				IssuesUpdateWorkflow(V1),
		),
		Check("Unknown with id and version",
			From(pipelinesv1.Unknown, PipelineId, V1).
				To(pipelinesv1.Updating, PipelineId, V1).
				IssuesUpdateWorkflow(V1),
		),
		Check("Creation succeeds",
			From(pipelinesv1.Creating, "", V1).
				WithWorkFlow(
					setWorkflowOutput(
						createWorkflow(CreateOperationLabel, argo.WorkflowSucceeded),
						WorkflowFactoryConstants.pipelineIdParameterName, PipelineId),
				).
				To(pipelinesv1.Succeeded, PipelineId, V1).
				DeletesAllWorkflows(),
		),
		Check("Creation succeeds with existing Id",
			From(pipelinesv1.Creating, AnotherPipelineId, V1).
				WithWorkFlow(
					setWorkflowOutput(
						createWorkflow(CreateOperationLabel, argo.WorkflowSucceeded),
						WorkflowFactoryConstants.pipelineIdParameterName, PipelineId),
				).
				To(pipelinesv1.Succeeded, PipelineId, V1).
				DeletesAllWorkflows(),
		),
		Check("Creation fails with Id",
			From(pipelinesv1.Creating, "", V1).
				WithWorkFlow(setWorkflowOutput(
					createWorkflow(CreateOperationLabel, argo.WorkflowFailed),
					WorkflowFactoryConstants.pipelineIdParameterName, PipelineId),
				).
				To(pipelinesv1.Failed, PipelineId, V1).
				DeletesAllWorkflows(),
		),
		Check("Creation fails",
			From(pipelinesv1.Creating, "", V1).
				WithWorkFlow(createWorkflow(CreateOperationLabel, argo.WorkflowFailed)).
				To(pipelinesv1.Failed, "", V1).
				DeletesAllWorkflows(),
		),
		Check("Creating without version",
			From(pipelinesv1.Creating, "", "").
				To(pipelinesv1.Failed, "", ""),
		),
		Check("Succeeded no update",
			From(pipelinesv1.Succeeded, PipelineId, V1),
		),
		Check("Succeeded with update",
			From(pipelinesv1.Succeeded, PipelineId, V0).
				To(pipelinesv1.Updating, PipelineId, V1).
				IssuesUpdateWorkflow(V1),
		),
		Check("Succeeded with update but no Id",
			From(pipelinesv1.Succeeded, "", V0).
				To(pipelinesv1.Creating, "", V1).
				IssuesCreationWorkflow(V1),
		),
		Check("Succeeded with update but no Id and no version",
			From(pipelinesv1.Succeeded, "", "").
				To(pipelinesv1.Creating, "", V1).
				IssuesCreationWorkflow(V1),
		),
		Check("Failed no update",
			From(pipelinesv1.Failed, PipelineId, V1),
		),
		Check("Failed with Update",
			From(pipelinesv1.Failed, PipelineId, V0).
				To(pipelinesv1.Updating, PipelineId, V1).
				IssuesUpdateWorkflow(V1),
		),
		Check("Failed with Update but no Id",
			From(pipelinesv1.Failed, "", V0).
				To(pipelinesv1.Creating, "", V1).
				IssuesCreationWorkflow(V1),
		),
		Check("Failed with Update but no Id and no version",
			From(pipelinesv1.Failed, "", "").
				To(pipelinesv1.Creating, "", V1).
				IssuesCreationWorkflow(V1),
		),
		Check("Updating succeeds",
			From(pipelinesv1.Updating, PipelineId, V1).
				WithWorkFlow(createWorkflow(UpdateOperationLabel, argo.WorkflowSucceeded)).
				To(pipelinesv1.Succeeded, PipelineId, V1).
				DeletesAllWorkflows(),
		),
		Check("Updating fails",
			From(pipelinesv1.Updating, PipelineId, V1).
				WithWorkFlow(createWorkflow(UpdateOperationLabel, argo.WorkflowFailed)).
				To(pipelinesv1.Failed, PipelineId, V1).
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
			From(pipelinesv1.Updating, "", V1).
				To(pipelinesv1.Failed, "", V1),
		),
		Check("updating without Id or version",
			From(pipelinesv1.Updating, "", "").
				To(pipelinesv1.Failed, "", ""),
		),
		Check("Deleting from Succeeded",
			From(pipelinesv1.Succeeded, PipelineId, V1).
				DeletionRequested().
				To(pipelinesv1.Deleting, PipelineId, V1).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Failed",
			From(pipelinesv1.Failed, PipelineId, V1).
				DeletionRequested().
				To(pipelinesv1.Deleting, PipelineId, V1).
				IssuesDeletionWorkflow(),
		),
		Check("Deletion succeeds",
			From(pipelinesv1.Deleting, PipelineId, V1).
				DeletionRequested().
				WithWorkFlow(createWorkflow(DeleteOperationLabel, argo.WorkflowSucceeded)).
				To(pipelinesv1.Deleted, PipelineId, V1).
				DeletesAllWorkflows(),
		),
		Check("Deletion fails",
			From(pipelinesv1.Deleting, PipelineId, V1).
				DeletionRequested().
				WithWorkFlow(createWorkflow(DeleteOperationLabel, argo.WorkflowFailed)).
				To(pipelinesv1.Deleting, PipelineId, V1).
				DeletesAllWorkflows(),
		),
		Check("Stay in deleted",
			From(pipelinesv1.Deleted, PipelineId, V1).
				IssuesCommand(DeletePipeline{}),
		),
	)
})
