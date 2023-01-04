package pipelines

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var WorkflowConstants = struct {
	OwnerKindLabelKey               string
	OwnerNameLabelKey               string
	OwnerNamespaceLabelKey          string
	OperationLabelKey               string
	CreateOperationLabel            string
	DeleteOperationLabel            string
	UpdateOperationLabel            string
	ConstructionFailedError         string
	ProviderNameParameterName       string
	ProviderOutputParameterName     string
	ResourceKindParameterName       string
	ResourceDefinitionParameterName string
	ResourceIdParameterName         string
}{
	OwnerKindLabelKey:               apis.Group + "/owner.kind",
	OwnerNameLabelKey:               apis.Group + "/owner.name",
	OwnerNamespaceLabelKey:          apis.Group + "/owner.namespace",
	OperationLabelKey:               apis.Group + "/operation",
	CreateOperationLabel:            "create",
	DeleteOperationLabel:            "delete",
	UpdateOperationLabel:            "update",
	ConstructionFailedError:         "error constructing workflow",
	ProviderNameParameterName:       "provider-name",
	ProviderOutputParameterName:     "provider-output",
	ResourceKindParameterName:       "resource-kind",
	ResourceDefinitionParameterName: "resource-definition",
	ResourceIdParameterName:         "resource-id",
}

type WorkflowFactory[R pipelinesv1.Resource] interface {
	ConstructCreationWorkflow(provider string, resource R) (*argo.Workflow, error)
	ConstructUpdateWorkflow(provider string, resource R) (*argo.Workflow, error)
	ConstructDeletionWorkflow(provider string, resource R) (*argo.Workflow, error)
}

type TemplateNameGenerator interface {
	CreateTemplate() string
	UpdateTemplate() string
	DeleteTemplate() string
}

type SuffixedTemplateNameGenerator struct {
	config config.Configuration
	suffix string
}

func CompiledTemplateNameGenerator(config config.Configuration) TemplateNameGenerator {
	return SuffixedTemplateNameGenerator{config: config, suffix: "compiled"}
}

func SimpleTemplateNameGenerator(config config.Configuration) TemplateNameGenerator {
	return SuffixedTemplateNameGenerator{config: config, suffix: "simple"}
}

func (ctng SuffixedTemplateNameGenerator) CreateTemplate() string {
	return fmt.Sprintf("%screate-%s", ctng.config.WorkflowTemplatePrefix, ctng.suffix)
}

func (ctng SuffixedTemplateNameGenerator) UpdateTemplate() string {
	return fmt.Sprintf("%supdate-%s", ctng.config.WorkflowTemplatePrefix, ctng.suffix)
}

func (ctng SuffixedTemplateNameGenerator) DeleteTemplate() string {
	return fmt.Sprintf("%sdelete", ctng.config.WorkflowTemplatePrefix)
}

type ResourceWorkflowFactory[R pipelinesv1.Resource] struct {
	Config                config.Configuration
	TemplateNameGenerator TemplateNameGenerator
	DefinitionCreator     func(R) (string, error)
}

func (workflows ResourceWorkflowFactory[R]) CommonWorkflowMeta(owner pipelinesv1.Resource, operation string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: fmt.Sprintf("%s-%s-", operation, owner.GetKind()),
		Namespace:    workflows.Config.WorkflowNamespace,
		Labels:       CommonWorkflowLabels(owner, operation),
	}
}

func CommonWorkflowLabels(owner pipelinesv1.Resource, operation string) map[string]string {
	return map[string]string{
		WorkflowConstants.OperationLabelKey:      operation,
		WorkflowConstants.OwnerKindLabelKey:      owner.GetKind(),
		WorkflowConstants.OwnerNameLabelKey:      owner.GetName(),
		WorkflowConstants.OwnerNamespaceLabelKey: owner.GetNamespace(),
	}
}

func (workflows *ResourceWorkflowFactory[R]) ConstructCreationWorkflow(provider string, resource R) (*argo.Workflow, error) {
	fmt.Println(workflows)
	fmt.Print(workflows.DefinitionCreator(resource))
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
						Name:  WorkflowConstants.ResourceKindParameterName,
						Value: argo.AnyStringPtr(resource.GetKind()),
					},
					{
						Name:  WorkflowConstants.ResourceDefinitionParameterName,
						Value: argo.AnyStringPtr(resourceDefinition),
					},
					{
						Name:  WorkflowConstants.ProviderNameParameterName,
						Value: argo.AnyStringPtr(provider),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.TemplateNameGenerator.CreateTemplate(),
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
						Name:  WorkflowConstants.ResourceKindParameterName,
						Value: argo.AnyStringPtr(resource.GetKind()),
					},
					{
						Name:  WorkflowConstants.ResourceDefinitionParameterName,
						Value: argo.AnyStringPtr(resourceDefinition),
					},
					{
						Name:  WorkflowConstants.ResourceIdParameterName,
						Value: argo.AnyStringPtr(resource.GetStatus().ProviderId.Id),
					},
					{
						Name:  WorkflowConstants.ProviderNameParameterName,
						Value: argo.AnyStringPtr(provider),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.TemplateNameGenerator.UpdateTemplate(),
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
						Name:  WorkflowConstants.ResourceKindParameterName,
						Value: argo.AnyStringPtr(resource.GetKind()),
					},
					{
						Name:  WorkflowConstants.ResourceIdParameterName,
						Value: argo.AnyStringPtr(resource.GetStatus().ProviderId.Id),
					},
					{
						Name:  WorkflowConstants.ProviderNameParameterName,
						Value: argo.AnyStringPtr(provider),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.TemplateNameGenerator.DeleteTemplate(),
			},
		},
	}, nil
}
