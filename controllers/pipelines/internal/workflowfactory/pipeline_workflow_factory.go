package workflowfactory

import (
	"fmt"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	config "github.com/sky-uk/kfp-operator/apis/config/v1beta1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1beta1"
	"github.com/sky-uk/kfp-operator/argo/common"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
)

type PipelineParamsCreator struct {
	Config config.KfpControllerConfigSpec
}

const defaultFramework = "default"

func (ppc PipelineParamsCreator) pipelineDefinition(
	pipeline *pipelinesv1.Pipeline,
) (providers.PipelineDefinition, error) {
	return providers.PipelineDefinition{
		Name: common.NamespacedName{
			Namespace: pipeline.ObjectMeta.Namespace,
			Name:      pipeline.ObjectMeta.Name,
		},
		Version:       pipeline.ComputeVersion(),
		Image:         pipeline.Spec.Image,
		Framework:     pipeline.Spec.Framework,
		TfxComponents: pipeline.Spec.TfxComponents,
		Env:           pipeline.Spec.Env,
		BeamArgs:      pipeline.Spec.BeamArgs,
	}, nil
}

func (ppc PipelineParamsCreator) additionalParams(pipeline *pipelinesv1.Pipeline) ([]argo.Parameter, error) {
	requestedFramework := defaultFramework
	if pipeline.Spec.Framework != "" {
		requestedFramework = pipeline.Spec.Framework
	}
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
) *ResourceWorkflowFactory[*pipelinesv1.Pipeline, providers.PipelineDefinition] {
	creator := PipelineParamsCreator{
		Config: config,
	}
	return &ResourceWorkflowFactory[*pipelinesv1.Pipeline, providers.PipelineDefinition]{
		DefinitionCreator:     creator.pipelineDefinition,
		WorkflowParamsCreator: creator.additionalParams,
		Config:                config,
		TemplateNameGenerator: CompiledTemplateNameGenerator(config),
	}
}
