package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
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

func (w WorkflowRepositoryImpl) SetupWithManager(mgr ctrl.Manager) error {
	return mgr.GetFieldIndexer().IndexField(context.Background(), &argo.Workflow{}, workflowOwnerKey, func(rawObj client.Object) []string {
		workflow := rawObj.(*argo.Workflow)

		owner := metav1.GetControllerOf(workflow)

		if owner == nil {
			return nil
		}

		isOwnedWorkflow := owner.APIVersion == apiGVStr && (owner.Kind == "Pipeline" || owner.Kind == "RunConfiguration")

		if !isOwnedWorkflow {
			return nil
		}

		return []string{owner.Name}
	})
}
