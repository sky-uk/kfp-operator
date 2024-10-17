package pipelines

import (
	"encoding/json"
	"fmt"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var WorkflowConstants = struct {
	OwnerKindLabelKey               string
	OwnerNameLabelKey               string
	OwnerNamespaceLabelKey          string
	ConstructionFailedError         string
	ProviderConfigParameterName     string
	ProviderNameParameterName       string
	ProviderOutputParameterName     string
	ResourceKindParameterName       string
	ResourceDefinitionParameterName string
	ResourceIdParameterName         string
}{
	OwnerKindLabelKey:               apis.Group + "/owner.kind",
	OwnerNameLabelKey:               apis.Group + "/owner.name",
	OwnerNamespaceLabelKey:          apis.Group + "/owner.namespace",
	ConstructionFailedError:         "error constructing workflow",
	ProviderConfigParameterName:     "provider-config",
	ProviderNameParameterName:       "provider-name",
	ProviderOutputParameterName:     "provider-output",
	ResourceKindParameterName:       "resource-kind",
	ResourceDefinitionParameterName: "resource-definition",
	ResourceIdParameterName:         "resource-id",
}

type WorkflowFactory[R pipelinesv1.Resource] interface {
	ConstructCreationWorkflow(provider pipelinesv1.Provider, resource R) (*argo.Workflow, error)
	ConstructUpdateWorkflow(provider pipelinesv1.Provider, resource R) (*argo.Workflow, error)
	ConstructDeletionWorkflow(provider pipelinesv1.Provider, resource R) (*argo.Workflow, error)
}

type TemplateNameGenerator interface {
	CreateTemplate() string
	UpdateTemplate() string
	DeleteTemplate() string
}

type SuffixedTemplateNameGenerator struct {
	config config.KfpControllerConfigSpec
	suffix string
}

func CompiledTemplateNameGenerator(config config.KfpControllerConfigSpec) TemplateNameGenerator {
	return SuffixedTemplateNameGenerator{config: config, suffix: "compiled"}
}

func SimpleTemplateNameGenerator(config config.KfpControllerConfigSpec) TemplateNameGenerator {
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

type ResourceWorkflowFactory[R pipelinesv1.Resource, ResourceDefinition any] struct {
	Config                config.KfpControllerConfigSpec
	TemplateNameGenerator TemplateNameGenerator
	DefinitionCreator     func(R) (ResourceDefinition, error)
}

func (workflows ResourceWorkflowFactory[R, ResourceDefinition]) CommonWorkflowMeta(owner pipelinesv1.Resource) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: fmt.Sprintf("%s-%s-", owner.GetKind(), owner.GetName()),
		Namespace:    workflows.Config.WorkflowNamespace,
		Labels:       CommonWorkflowLabels(owner),
	}
}

func CommonWorkflowLabels(owner pipelinesv1.Resource) map[string]string {
	return map[string]string{
		WorkflowConstants.OwnerKindLabelKey:      owner.GetKind(),
		WorkflowConstants.OwnerNameLabelKey:      owner.GetName(),
		WorkflowConstants.OwnerNamespaceLabelKey: owner.GetNamespace(),
	}
}

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) resourceDefinitionYaml(resource R) (string, error) {
	resourceDefinition, err := workflows.DefinitionCreator(resource)
	if err != nil {
		return "", err
	}

	marshalled, err := yaml.Marshal(&resourceDefinition)
	if err != nil {
		return "", err
	}

	return string(marshalled), nil
}

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructCreationWorkflow(provider pipelinesv1.Provider, resource R) (*argo.Workflow, error) {
	resourceDefinition, err := workflows.resourceDefinitionYaml(resource)
	if err != nil {
		return nil, err
	}

	providerConf, err := json.Marshal(provider.Spec)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *workflows.CommonWorkflowMeta(resource),
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
						Value: argo.AnyStringPtr(provider.Name),
					},
					{
						Name:  WorkflowConstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(string(providerConf)),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.TemplateNameGenerator.CreateTemplate(),
			},
		},
	}, nil
}

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructUpdateWorkflow(provider pipelinesv1.Provider, resource R) (*argo.Workflow, error) {
	resourceDefinition, err := workflows.resourceDefinitionYaml(resource)
	if err != nil {
		return nil, err
	}

	providerConf, err := json.Marshal(provider.Spec)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *workflows.CommonWorkflowMeta(resource),
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
						Value: argo.AnyStringPtr(provider.Name),
					},
					{
						Name:  WorkflowConstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(string(providerConf)),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.TemplateNameGenerator.UpdateTemplate(),
			},
		},
	}, nil
}

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructDeletionWorkflow(provider pipelinesv1.Provider, resource R) (*argo.Workflow, error) {
	providerConf, err := json.Marshal(provider.Spec)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *workflows.CommonWorkflowMeta(resource),
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
						Value: argo.AnyStringPtr(provider.Name),
					},
					{
						Name:  WorkflowConstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(string(providerConf)),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.TemplateNameGenerator.DeleteTemplate(),
			},
		},
	}, nil
}
