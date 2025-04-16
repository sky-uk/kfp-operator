package workflowfactory

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/apis"
	"strings"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
)

type PipelineParamsCreator struct{}

func findFramework(provider pipelineshub.Provider, pipeline *pipelineshub.Pipeline) (*pipelineshub.Framework, bool) {
	requestedFramework := strings.ToLower(pipeline.Spec.Framework.Type)

	return apis.Find(provider.Spec.Frameworks, func(framework pipelineshub.Framework) bool {
		return framework.Name == requestedFramework
	})
}

func (ppc PipelineParamsCreator) pipelineDefinition(
	provider pipelineshub.Provider, pipeline *pipelineshub.Pipeline,
) ([]pipelineshub.Patch, providers.PipelineDefinition, error) {
	framework, found := findFramework(provider, pipeline)

	if !found {
		return nil, providers.PipelineDefinition{}, &workflowconstants.WorkflowParameterError{SubError: fmt.Sprintf("[%s] framework not support by provider", pipeline.Spec.Framework.Type)}
	}

	return framework.Patches, providers.PipelineDefinition{
		Name: common.NamespacedName{
			Namespace: pipeline.ObjectMeta.Namespace,
			Name:      pipeline.ObjectMeta.Name,
		},
		Version:   pipeline.ComputeVersion(),
		Image:     pipeline.Spec.Image,
		Framework: pipeline.Spec.Framework,
		Env:       pipeline.Spec.Env,
	}, nil
}

func (ppc PipelineParamsCreator) additionalParams(provider pipelineshub.Provider, pipeline *pipelineshub.Pipeline) ([]argo.Parameter, error) {
	framework, found := findFramework(provider, pipeline)

	if !found {
		return nil, &workflowconstants.WorkflowParameterError{SubError: fmt.Sprintf("[%s] framework not support by provider", pipeline.Spec.Framework.Type)}
	}

	return []argo.Parameter{
		{
			Name:  workflowconstants.PipelineFrameworkImageParameterName,
			Value: argo.AnyStringPtr(framework.Image),
		},
	}, nil
}

func PipelineWorkflowFactory(
	config config.KfpControllerConfigSpec,
) *ResourceWorkflowFactory[*pipelineshub.Pipeline, providers.PipelineDefinition] {
	creator := PipelineParamsCreator{}
	return &ResourceWorkflowFactory[*pipelineshub.Pipeline, providers.PipelineDefinition]{
		DefinitionCreator:     creator.pipelineDefinition,
		WorkflowParamsCreator: creator.additionalParams,
		Config:                config,
		TemplateNameGenerator: CompiledTemplateNameGenerator(config),
	}
}
