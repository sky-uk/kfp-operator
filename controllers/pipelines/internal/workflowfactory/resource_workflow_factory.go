package workflowfactory

import (
	"encoding/json"
	"fmt"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	. "github.com/sky-uk/kfp-operator/controllers/pipelines/internal/jsonutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkflowFactory[R pipelineshub.Resource] interface {
	ConstructCreationWorkflow(
		provider pipelineshub.Provider,
		providerSvc corev1.Service,
		resource R,
	) (*argo.Workflow, error)

	ConstructUpdateWorkflow(
		provider pipelineshub.Provider,
		providerSvc corev1.Service,
		resource R,
	) (*argo.Workflow, error)

	ConstructDeletionWorkflow(
		provider pipelineshub.Provider,
		providerSvc corev1.Service,
		resource R,
	) (*argo.Workflow, error)
}

type ResourceWorkflowFactory[R pipelineshub.Resource, ResourceDefinition any] struct {
	Config                config.KfpControllerConfigSpec
	TemplateSuffix        string
	DefinitionCreator     func(pipelineshub.Provider, R) ([]pipelineshub.Patch, ResourceDefinition, error)
	WorkflowParamsCreator func(pipelineshub.Provider, R) ([]argo.Parameter, error)
	workflowBuilder       *BaseWorkflowBuilder
}

const (
	CompiledSuffix = "compiled"
	SimpleSuffix   = "simple"
)

// Template name methods - create and update operations use suffixed templates to differentiate
// between resource types (e.g., "compiled" for pipelines, "simple" for basic resources).
// Delete operations use a single shared template since deletion is generic across all resource types.
func (rwf ResourceWorkflowFactory[R, ResourceDefinition]) createTemplateName() string {
	return fmt.Sprintf("%screate-%s", rwf.Config.WorkflowTemplatePrefix, rwf.TemplateSuffix)
}

func (rwf ResourceWorkflowFactory[R, ResourceDefinition]) updateTemplateName() string {
	return fmt.Sprintf("%supdate-%s", rwf.Config.WorkflowTemplatePrefix, rwf.TemplateSuffix)
}

func (rwf ResourceWorkflowFactory[R, ResourceDefinition]) deleteTemplateName() string {
	return fmt.Sprintf("%sdelete", rwf.Config.WorkflowTemplatePrefix)
}

func WorkflowParamsCreatorNoop[R any](provider pipelineshub.Provider, _ R) ([]argo.Parameter, error) {
	return []argo.Parameter{}, nil
}

func NewResourceWorkflowFactory[R pipelineshub.Resource, ResourceDefinition any](
	config config.KfpControllerConfigSpec,
	templateSuffix string,
	definitionCreator func(pipelineshub.Provider, R) ([]pipelineshub.Patch, ResourceDefinition, error),
	workflowParamsCreator func(pipelineshub.Provider, R) ([]argo.Parameter, error),
) *ResourceWorkflowFactory[R, ResourceDefinition] {
	return &ResourceWorkflowFactory[R, ResourceDefinition]{
		Config:                config,
		TemplateSuffix:        templateSuffix,
		DefinitionCreator:     definitionCreator,
		WorkflowParamsCreator: workflowParamsCreator,
		workflowBuilder:       NewBaseWorkflowBuilder(config),
	}
}

// CommonWorkflowMeta is deprecated, use workflowBuilder.commonWorkflowMeta instead
func (rwf ResourceWorkflowFactory[R, ResourceDefinition]) CommonWorkflowMeta(
	owner pipelineshub.Resource,
) *metav1.ObjectMeta {
	return rwf.workflowBuilder.commonWorkflowMeta(owner)
}

func (rwf *ResourceWorkflowFactory[R, ResourceDefinition]) resourceDefinitionJson(provider pipelineshub.Provider, resource R) (string, error) {
	patches, resourceDefinition, err := rwf.DefinitionCreator(provider, resource)
	if err != nil {
		return "", err
	}

	marshalled, err := json.Marshal(&resourceDefinition)
	if err != nil {
		return "", err
	}

	patchedJsonString, err := PatchJson(patches, marshalled)
	if err != nil {
		return "", err
	}

	return patchedJsonString, nil
}

func (rwf *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructCreationWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
) (*argo.Workflow, error) {
	resourceDefinition, err := rwf.resourceDefinitionJson(provider, resource)
	if err != nil {
		return nil, err
	}

	baseParams := rwf.workflowBuilder.BuildCreationParams(resourceDefinition)

	additionalParams, err := rwf.WorkflowParamsCreator(provider, resource)
	if err != nil {
		return nil, err
	}

	return rwf.workflowBuilder.BuildWorkflow(
		resource,
		provider,
		providerSvc,
		rwf.createTemplateName(),
		baseParams,
		additionalParams,
	)
}

func (rwf *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructUpdateWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
) (*argo.Workflow, error) {
	resourceDefinition, err := rwf.resourceDefinitionJson(provider, resource)
	if err != nil {
		return nil, err
	}

	baseParams := rwf.workflowBuilder.BuildUpdateParams(
		resourceDefinition,
		resource.GetStatus().Provider.Id,
	)

	additionalParams, err := rwf.WorkflowParamsCreator(provider, resource)
	if err != nil {
		return nil, err
	}

	return rwf.workflowBuilder.BuildWorkflow(
		resource,
		provider,
		providerSvc,
		rwf.updateTemplateName(),
		baseParams,
		additionalParams,
	)
}

func (rwf *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructDeletionWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
) (*argo.Workflow, error) {
	baseParams := rwf.workflowBuilder.BuildDeletionParams(resource.GetStatus().Provider.Id)

	return rwf.workflowBuilder.BuildWorkflow(
		resource,
		provider,
		providerSvc,
		rwf.deleteTemplateName(),
		baseParams,
		[]argo.Parameter{},
	)
}
