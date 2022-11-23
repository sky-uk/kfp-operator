//go:build unit
// +build unit

package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StubbedWorkflows struct {
	Workflows []argo.Workflow
}

func (sw StubbedWorkflows) GetByLabels(_ context.Context, _ map[string]string) []argo.Workflow {
	return sw.Workflows
}

func (sw StubbedWorkflows) CreateWorkflowForResource(_ context.Context, _ *argo.Workflow, _ pipelinesv1.Resource) error {
	return nil
}

func (sw StubbedWorkflows) MarkWorkflowAsProcessed(_ context.Context, _ *argo.Workflow) error {
	return nil
}

func (sw *StubbedWorkflows) AddWorkflow(workflow argo.Workflow) {
	sw.Workflows = append(sw.Workflows, workflow)
}

func CreateTestWorkflow(operation string, phase argo.WorkflowPhase) *argo.Workflow {
	return &argo.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operation,
			Namespace: "default",
			Labels: map[string]string{
				WorkflowConstants.OperationLabelKey: operation,
			},
		},
		Status: argo.WorkflowStatus{
			Phase: phase,
		},
	}
}
