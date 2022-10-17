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

type ExperimentStateTransitionTestCase struct {
	workflowFactory WorkflowFactory[*pipelinesv1.Experiment]
	Experiment      *pipelinesv1.Experiment
	SystemStatus    StubbedWorkflows
	Commands        []Command
}

type FailingExperimentWorkflowFactory struct{}

func (f FailingExperimentWorkflowFactory) ConstructCreationWorkflow(_ *pipelinesv1.Experiment) (*argo.Workflow, error) {
	return nil, fmt.Errorf("an error occurred")
}

func (f FailingExperimentWorkflowFactory) ConstructUpdateWorkflow(_ *pipelinesv1.Experiment) (*argo.Workflow, error) {
	return nil, fmt.Errorf("an error occurred")
}

func (f FailingExperimentWorkflowFactory) ConstructDeletionWorkflow(_ *pipelinesv1.Experiment) (*argo.Workflow, error) {
	return nil, fmt.Errorf("an error occurred")
}

func (st ExperimentStateTransitionTestCase) WorkflowConstructionFails() ExperimentStateTransitionTestCase {
	st.workflowFactory = FailingExperimentWorkflowFactory{}
	return st
}

func (st ExperimentStateTransitionTestCase) WithWorkFlow(workflow *argo.Workflow) ExperimentStateTransitionTestCase {
	st.SystemStatus.AddWorkflow(*workflow)
	return st
}

func (st ExperimentStateTransitionTestCase) WithCreateWorkFlow(phase argo.WorkflowPhase) ExperimentStateTransitionTestCase {
	return st.WithWorkFlow(CreateTestWorkflow(WorkflowConstants.CreateOperationLabel, phase))
}

func (st ExperimentStateTransitionTestCase) WithSucceededCreateWorkFlow(kfpId string, providerError string) ExperimentStateTransitionTestCase {
	return st.WithWorkFlow(
		setProviderOutput(
			CreateTestWorkflow(WorkflowConstants.CreateOperationLabel, argo.WorkflowSucceeded),
			base.Output{Id: kfpId, ProviderError: providerError},
		),
	)
}

func (st ExperimentStateTransitionTestCase) WithFailedUpdateWorkflow() ExperimentStateTransitionTestCase {
	return st.WithWorkFlow(
		CreateTestWorkflow(WorkflowConstants.UpdateOperationLabel, argo.WorkflowFailed),
	)
}

func (st ExperimentStateTransitionTestCase) WithSucceededUpdateWorkflow(kfpId string, providerError string) ExperimentStateTransitionTestCase {
	return st.WithWorkFlow(
		setProviderOutput(
			CreateTestWorkflow(WorkflowConstants.UpdateOperationLabel, argo.WorkflowSucceeded),
			base.Output{Id: kfpId, ProviderError: providerError},
		),
	)
}

func (st ExperimentStateTransitionTestCase) WithDeletionWorkflow(phase argo.WorkflowPhase) ExperimentStateTransitionTestCase {
	return st.WithWorkFlow(
		CreateTestWorkflow(WorkflowConstants.DeleteOperationLabel, phase),
	)
}

func (st ExperimentStateTransitionTestCase) WithSucceededDeletionWorkflow(kfpId string, providerError string) ExperimentStateTransitionTestCase {
	return st.WithWorkFlow(
		setProviderOutput(
			CreateTestWorkflow(WorkflowConstants.DeleteOperationLabel, argo.WorkflowSucceeded),
			base.Output{Id: kfpId, ProviderError: providerError},
		),
	)
}

func (st ExperimentStateTransitionTestCase) IssuesCreationWorkflow() ExperimentStateTransitionTestCase {
	creationWorkflow, _ := st.workflowFactory.ConstructCreationWorkflow(st.Experiment)
	return st.IssuesCommand(CreateWorkflow{Workflow: *creationWorkflow})
}

func (st ExperimentStateTransitionTestCase) IssuesUpdateWorkflow() ExperimentStateTransitionTestCase {
	updateWorkflow, _ := st.workflowFactory.ConstructUpdateWorkflow(st.Experiment)
	return st.IssuesCommand(CreateWorkflow{Workflow: *updateWorkflow})
}

func (st ExperimentStateTransitionTestCase) IssuesDeletionWorkflow() ExperimentStateTransitionTestCase {
	deletionWorkflow, _ := st.workflowFactory.ConstructDeletionWorkflow(st.Experiment)
	return st.IssuesCommand(CreateWorkflow{Workflow: *deletionWorkflow})
}

func (st ExperimentStateTransitionTestCase) DeletesAllWorkflows() ExperimentStateTransitionTestCase {
	return st.IssuesCommand(DeleteWorkflows{
		Workflows: st.SystemStatus.Workflows,
	})
}

func (st ExperimentStateTransitionTestCase) AcquireExperiment() ExperimentStateTransitionTestCase {
	return st.IssuesCommand(AcquireResource{})
}

func (st ExperimentStateTransitionTestCase) ReleaseExperiment() ExperimentStateTransitionTestCase {
	return st.IssuesCommand(ReleaseResource{})
}

func (st ExperimentStateTransitionTestCase) IssuesCommand(command Command) ExperimentStateTransitionTestCase {
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
		WorkflowFactoryBase: WorkflowFactoryBase{
			Config: config.Configuration{},
		},
	}

	kfpId := "12345"
	anotherKfpId := "67890"
	providerError := "a provider error has occurred"
	specv1 := pipelinesv1.RandomExperimentSpec()
	v0 := pipelinesv1.ExperimentSpec{}.ComputeVersion()
	v1 := specv1.ComputeVersion()
	UnknownState := apis.SynchronizationState(apis.RandomString())

	var Check = func(description string, transition ExperimentStateTransitionTestCase) TableEntry {
		return Entry(
			description,
			transition,
		)
	}

	var From = func(status apis.SynchronizationState, id string, version string) ExperimentStateTransitionTestCase {
		experiment := pipelinesv1.RandomExperiment()
		experiment.Spec = specv1
		experiment.Status = apis.Status{
			SynchronizationState: status,
			Version:              version,
			KfpId:                id,
		}

		return ExperimentStateTransitionTestCase{
			workflowFactory: workflowFactory,
			Experiment:      experiment,
			Commands:        []Command{},
		}
	}

	DescribeTable("State transitions", func(st ExperimentStateTransitionTestCase) {
		var stateHandler = ExperimentStateHandler{
			WorkflowRepository: st.SystemStatus,
			WorkflowFactory:    st.workflowFactory,
		}
		commands := stateHandler.stateTransition(context.Background(), st.Experiment)
		Expect(commands).To(Equal(st.Commands))
	},
		Check("Empty",
			From(UnknownState, "", "").
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Empty and workflow creation fails",
			From(UnknownState, "", "").
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Empty with version",
			From(UnknownState, "", v1).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Empty with id",
			From(UnknownState, kfpId, "").
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Empty with id and workflow creation fails",
			From(UnknownState, kfpId, "").
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Empty with id and version",
			From(UnknownState, kfpId, v1).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Creating succeeds",
			From(apis.Creating, "", v1).
				AcquireExperiment().
				WithSucceededCreateWorkFlow(kfpId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Succeeded).
					WithKfpId(kfpId).
					WithVersion(v1)).
				DeletesAllWorkflows(),
		),
		Check("Creating succeeds without kfpId or provider error",
			From(apis.Creating, kfpId, v1).
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
			From(apis.Creating, anotherKfpId, v1).
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
			From(apis.Creating, "", v1).
				AcquireExperiment().
				WithCreateWorkFlow(argo.WorkflowFailed).
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithVersion(v1).
					WithMessage("operation failed")).
				DeletesAllWorkflows(),
		),
		Check("Creating without version",
			From(apis.Creating, "", "").
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithMessage("creating experiment with empty version")),
		),
		Check("Succeeded no update",
			From(apis.Succeeded, kfpId, v1).
				AcquireExperiment(),
		),
		Check("Succeeded with update",
			From(apis.Succeeded, kfpId, v0).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Succeeded with update and workflow creation fails",
			From(apis.Succeeded, kfpId, v0).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Succeeded with update but no KfpId",
			From(apis.Succeeded, "", v0).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Succeeded with update and workflow creation fails",
			From(apis.Succeeded, "", v0).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Succeeded with update but no KfpId and no version",
			From(apis.Succeeded, "", "").
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Failed no update",
			From(apis.Failed, kfpId, v1).
				AcquireExperiment(),
		),
		Check("Failed with update",
			From(apis.Failed, kfpId, v0).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Updating).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesUpdateWorkflow(),
		),
		Check("Failed with update and workflow creation fails",
			From(apis.Failed, kfpId, v0).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithKfpId(kfpId).
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Failed with update but no KfpId",
			From(apis.Failed, "", v0).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Failed with update but no KfpId and workflow creation fails",
			From(apis.Failed, "", v0).
				AcquireExperiment().
				WorkflowConstructionFails().
				IssuesCommand(*NewSetStatus().
					WithVersion(v1).
					WithMessage(WorkflowConstants.ConstructionFailedError).
					WithSynchronizationState(apis.Failed)),
		),
		Check("Failed with update but no KfpId and no version",
			From(apis.Failed, "", "").
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Creating).
					WithVersion(v1)).
				IssuesCreationWorkflow(),
		),
		Check("Updating succeeds with kfpId",
			From(apis.Updating, kfpId, v1).
				AcquireExperiment().
				WithSucceededUpdateWorkflow(anotherKfpId, "").
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Succeeded).
					WithKfpId(anotherKfpId).
					WithVersion(v1)).
				DeletesAllWorkflows(),
		),
		Check("Updating succeeds without kfpId or provider error",
			From(apis.Updating, kfpId, v1).
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
			From(apis.Updating, kfpId, v1).
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
			From(apis.Updating, kfpId, v1).
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
			From(apis.Updating, kfpId, "").
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithKfpId(kfpId).
					WithMessage("updating experiment with empty version or kfpId")),
		),
		Check("Updating without KfpId",
			From(apis.Updating, "", v1).
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithVersion(v1).
					WithMessage("updating experiment with empty version or kfpId")),
		),
		Check("Updating without KfpId or version",
			From(apis.Updating, "", "").
				AcquireExperiment().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Failed).
					WithMessage("updating experiment with empty version or kfpId")),
		),
		Check("Deleting from Succeeded",
			From(apis.Succeeded, kfpId, v1).
				AcquireExperiment().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Succeeded without kfpId",
			From(apis.Succeeded, "", v1).
				AcquireExperiment().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleted).
					WithVersion(v1)),
		),
		Check("Deleting from Failed",
			From(apis.Failed, kfpId, v1).
				AcquireExperiment().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleting).
					WithKfpId(kfpId).
					WithVersion(v1)).
				IssuesDeletionWorkflow(),
		),
		Check("Deleting from Failed without kfpId",
			From(apis.Failed, "", v1).
				AcquireExperiment().
				DeletionRequested().
				IssuesCommand(*NewSetStatus().
					WithSynchronizationState(apis.Deleted).
					WithVersion(v1)),
		),
		Check("Deletion succeeds with kfpId",
			From(apis.Deleting, kfpId, v1).
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
			From(apis.Deleting, kfpId, v1).
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
			From(apis.Deleting, kfpId, v1).
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
			From(apis.Deleting, kfpId, v1).
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
			From(apis.Deleted, kfpId, v1).
				ReleaseExperiment(),
		))
})
