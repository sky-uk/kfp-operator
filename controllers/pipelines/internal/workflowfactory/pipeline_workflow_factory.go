package workflowfactory

import (
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
)

type PipelineDefinitionCreator struct {
	Config config.KfpControllerConfigSpec
}

func (pdc PipelineDefinitionCreator) pipelineDefinition(
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

func PipelineWorkflowFactory(
	config config.KfpControllerConfigSpec,
) *ResourceWorkflowFactory[*pipelinesv1.Pipeline, providers.PipelineDefinition] {
	return &ResourceWorkflowFactory[*pipelinesv1.Pipeline, providers.PipelineDefinition]{
		DefinitionCreator: PipelineDefinitionCreator{
			Config: config,
		}.pipelineDefinition,
		Config:                config,
		TemplateNameGenerator: CompiledTemplateNameGenerator(config),
	}
}
