package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/controllers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type WorkflowRepository interface {
	GetByOperation(ctx context.Context, operation string, namespacedName types.NamespacedName, label string) []argo.Workflow
}

type WorkflowRepositoryImpl struct {
	Client controllers.OptInClient
}

func (w WorkflowRepositoryImpl) GetByOperation(ctx context.Context, operation string, namespacedName types.NamespacedName, label string) []argo.Workflow {
	logger := log.FromContext(ctx)
	var workflows argo.WorkflowList

	if err := w.Client.NonCached.List(ctx, &workflows, client.InNamespace(namespacedName.Namespace), client.MatchingLabels{PipelineWorkflowConstants.OperationLabelKey: operation, label: namespacedName.Name}); err != nil {
		//TODO: errors should be propagated to the caller
		logger.V(3).Error(err, "no matching workflows")
	} else {
		logger.V(3).Info("matching workflows", "workflows", workflows.Items)
	}

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
