package workflowfactory

import (
	"fmt"
	"strings"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
)

type PipelineParamsCreator struct {
	Config config.KfpControllerConfigSpec
}

const defaultFramework = "default"

func (ppc PipelineParamsCreator) pipelineDefinition(
	pipeline *pipelineshub.Pipeline,
) (providers.PipelineDefinition, error) {
	return providers.PipelineDefinition{
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

func (ppc PipelineParamsCreator) additionalParams(pipeline *pipelineshub.Pipeline) ([]argo.Parameter, error) {
	requestedFramework := strings.ToLower(pipeline.Spec.Framework.Type)
	frameworkImage, found := ppc.Config.PipelineFrameworkImages[requestedFramework]
	if !found {
		return nil, &workflowconstants.WorkflowParameterError{SubError: fmt.Sprintf("[%s] framework not found", requestedFramework)}
	}

	return []argo.Parameter{
		{
			Name:  workflowconstants.PipelineFrameworkImageParameterName,
			Value: argo.AnyStringPtr(frameworkImage),
		},
	}, nil
}

func PipelineWorkflowFactory(
	config config.KfpControllerConfigSpec,
) *ResourceWorkflowFactory[*pipelineshub.Pipeline, providers.PipelineDefinition] {
	creator := PipelineParamsCreator{
		Config: config,
	}
	return &ResourceWorkflowFactory[*pipelineshub.Pipeline, providers.PipelineDefinition]{
		DefinitionCreator:     creator.pipelineDefinition,
		WorkflowParamsCreator: creator.additionalParams,
		Config:                config,
		TemplateNameGenerator: CompiledTemplateNameGenerator(config),
	}
}
