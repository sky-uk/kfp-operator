package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkflowRepository interface {
	GetByOperation(ctx context.Context, operation string, namespacedName types.NamespacedName, label string) []argo.Workflow
}

type WorkflowRepositoryImpl struct {
	client.Client
}

func (w WorkflowRepositoryImpl) GetByOperation(ctx context.Context, operation string, namespacedName types.NamespacedName, label string) []argo.Workflow {
	var workflows argo.WorkflowList

	w.List(ctx, &workflows, client.InNamespace(namespacedName.Namespace), client.MatchingLabels{PipelineWorkflowConstants.OperationLabelKey: operation, label: namespacedName.Name})

	return workflows.Items
}
