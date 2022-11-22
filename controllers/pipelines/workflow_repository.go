package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"github.com/sky-uk/kfp-operator/controllers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var WorkflowRepositoryConstants = struct {
	WorkflowProcessedLabel string
}{
	WorkflowProcessedLabel: pipelinesv1.GroupVersion.Group + "/processed",
}

type WorkflowRepository interface {
	CreateWorkflowForResource(ctx context.Context, workflow *argo.Workflow, resource pipelinesv1.Resource) error
	GetByLabels(ctx context.Context, namespace string, matchingLabels map[string]string) []argo.Workflow
	MarkWorkflowAsProcessed(ctx context.Context, workflow *argo.Workflow) error
}

type WorkflowRepositoryImpl struct {
	Client controllers.OptInClient
	Config config.Configuration
	Scheme *runtime.Scheme
}

func (w WorkflowRepositoryImpl) CreateWorkflowForResource(ctx context.Context, workflow *argo.Workflow, resource pipelinesv1.Resource) error {
	if err := ctrl.SetControllerReference(resource, workflow, w.Scheme); err != nil {
		return err
	}

	return w.Client.Create(ctx, workflow)
}

func (w WorkflowRepositoryImpl) GetByLabels(ctx context.Context, namespace string, matchingLabels map[string]string) []argo.Workflow {
	logger := log.FromContext(ctx)
	var workflows argo.WorkflowList

	sel := labels.NewSelector()

	req, err := labels.NewRequirement(WorkflowRepositoryConstants.WorkflowProcessedLabel, selection.DoesNotExist, []string{})
	if err != nil {
		return []argo.Workflow{}
	}
	sel = sel.Add(*req)

	for label, value := range matchingLabels {
		req, err = labels.NewRequirement(label, selection.Equals, []string{value})
		if err != nil {
			return []argo.Workflow{}
		}
		sel = sel.Add(*req)
	}

	if err := w.Client.NonCached.List(ctx, &workflows, client.InNamespace(namespace), client.MatchingLabelsSelector{Selector: sel}); err != nil {
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

func (w WorkflowRepositoryImpl) MarkWorkflowAsProcessed(ctx context.Context, workflow *argo.Workflow) error {
	logger := log.FromContext(ctx)

	logger.V(1).Info("marking child workflow as processed", LogKeys.Workflow, workflow)
	workflowLabels := workflow.GetLabels()
	if workflowLabels == nil {
		workflowLabels = map[string]string{}
	}
	workflowLabels[WorkflowRepositoryConstants.WorkflowProcessedLabel] = "true"
	workflow.SetLabels(workflowLabels)

	return w.Client.Update(ctx, workflow)
}
