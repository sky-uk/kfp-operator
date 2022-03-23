//go:build unit
// +build unit

package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type StubbedWorkflows struct {
	Workflows []argo.Workflow
}

func (sw StubbedWorkflows) GetByLabels(_ context.Context, _ types.NamespacedName, _ map[string]string) []argo.Workflow {
	return sw.Workflows
}

func (sw StubbedWorkflows) CreateWorkflow(ctx context.Context, workflow *argo.Workflow) error {
	return nil
}

func (sw StubbedWorkflows) DeleteWorkflow(ctx context.Context, workflow *argo.Workflow) error {
	return nil
}

func (sw *StubbedWorkflows) AddWorkflow(workflow argo.Workflow) {
	sw.Workflows = append(sw.Workflows, workflow)
}

func (sw *StubbedWorkflows) CreateTestWorkflow(operation string, phase argo.WorkflowPhase) *argo.Workflow {
	return &argo.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operation,
			Namespace: "default",
			Labels: map[string]string{
				PipelineWorkflowConstants.OperationLabelKey: operation,
			},
		},
		Status: argo.WorkflowStatus{
			Phase: phase,
		},
	}
}
