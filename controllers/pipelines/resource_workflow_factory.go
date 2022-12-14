package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
)

var ResourceWorkflowConstants = struct {
	ResourceKindParameterName       string
	ResourceDefinitionParameterName string
	ResourceIdParameterName         string
}{
	ResourceKindParameterName:       "resource-kind",
	ResourceDefinitionParameterName: "resource-definition",
	ResourceIdParameterName:         "resource-id",
}

type ResourceWorkflowFactory[R pipelinesv1.Resource] struct {
	WorkflowFactoryBase
	DefinitionCreator func(R) (string, error)
}

func (workflows *ResourceWorkflowFactory[R]) ConstructCreationWorkflow(provider string, resource R) (*argo.Workflow, error) {
	resourceDefinition, err := workflows.DefinitionCreator(resource)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *workflows.CommonWorkflowMeta(resource, WorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  ResourceWorkflowConstants.ResourceKindParameterName,
						Value: argo.AnyStringPtr(resource.GetKind()),
					},
					{
						Name:  ResourceWorkflowConstants.ResourceDefinitionParameterName,
						Value: argo.AnyStringPtr(resourceDefinition),
					},
					{
						Name:  WorkflowConstants.ProviderNameParameterName,
						Value: argo.AnyStringPtr(provider),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.Config.WorkflowTemplatePrefix + "create",
			},
		},
	}, nil
}

func (workflows *ResourceWorkflowFactory[R]) ConstructUpdateWorkflow(provider string, resource R) (*argo.Workflow, error) {
	resourceDefinition, err := workflows.DefinitionCreator(resource)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *workflows.CommonWorkflowMeta(resource, WorkflowConstants.UpdateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  ResourceWorkflowConstants.ResourceKindParameterName,
						Value: argo.AnyStringPtr(resource.GetKind()),
					},
					{
						Name:  ResourceWorkflowConstants.ResourceDefinitionParameterName,
						Value: argo.AnyStringPtr(resourceDefinition),
					},
					{
						Name:  ResourceWorkflowConstants.ResourceIdParameterName,
						Value: argo.AnyStringPtr(resource.GetStatus().ProviderId.Id),
					},
					{
						Name:  WorkflowConstants.ProviderNameParameterName,
						Value: argo.AnyStringPtr(provider),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.Config.WorkflowTemplatePrefix + "update",
			},
		},
	}, nil
}

func (workflows *ResourceWorkflowFactory[R]) ConstructDeletionWorkflow(provider string, resource R) (*argo.Workflow, error) {
	return &argo.Workflow{
		ObjectMeta: *workflows.CommonWorkflowMeta(resource, WorkflowConstants.DeleteOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  ResourceWorkflowConstants.ResourceKindParameterName,
						Value: argo.AnyStringPtr(resource.GetKind()),
					},
					{
						Name:  ResourceWorkflowConstants.ResourceIdParameterName,
						Value: argo.AnyStringPtr(resource.GetStatus().ProviderId.Id),
					},
					{
						Name:  WorkflowConstants.ProviderNameParameterName,
						Value: argo.AnyStringPtr(provider),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.Config.WorkflowTemplatePrefix + "delete",
			},
		},
	}, nil
}
