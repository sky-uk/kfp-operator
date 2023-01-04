package pipelines

import (
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	providers "github.com/sky-uk/kfp-operator/providers/base"
	"gopkg.in/yaml.v2"
)

type PipelineDefinitionCreator struct {
	Config config.Configuration
}

func NamedValuesToMap(namedValues []apis.NamedValue) map[string]string {
	m := make(map[string]string)

	for _, nv := range namedValues {
		m[nv.Name] = nv.Value
	}

	return m
}

func NamedValuesToMultiMap(namedValues []apis.NamedValue) map[string][]string {
	multimap := make(map[string][]string)

	for _, nv := range namedValues {
		if _, found := multimap[nv.Name]; !found {
			multimap[nv.Name] = []string{}
		}

		multimap[nv.Name] = append(multimap[nv.Name], nv.Value)
	}

	return multimap
}

func (pdc PipelineDefinitionCreator) pipelineDefinitionYaml(pipeline *pipelinesv1.Pipeline) (string, error) {
	// TODO: should come from config
	servingPath := "/serving"
	tempPath := "/tmp"

	pipelineRoot := pdc.Config.PipelineStorage + "/" + pipeline.ObjectMeta.Name

	beamArgs := append(pdc.Config.DefaultBeamArgs, pipeline.Spec.BeamArgs...)
	beamArgs = append(beamArgs, apis.NamedValue{Name: "temp_location", Value: pipelineRoot + tempPath})

	pipelineDefinition := providers.PipelineDefinition{
		RootLocation:    pipelineRoot,
		ServingLocation: pipelineRoot + servingPath,
		Name:            pipeline.ObjectMeta.Name,
		Version:         pipeline.ComputeVersion(),
		Image:           pipeline.Spec.Image,
		TfxComponents:   pipeline.Spec.TfxComponents,
		Env:             NamedValuesToMap(pipeline.Spec.Env),
		BeamArgs:        NamedValuesToMultiMap(beamArgs),
	}

	marshalled, err := yaml.Marshal(&pipelineDefinition)
	if err != nil {
		return "", err
	}

	return string(marshalled), nil
}

func PipelineWorkflowFactory(config config.Configuration) ResourceWorkflowFactory[*pipelinesv1.Pipeline] {
	return ResourceWorkflowFactory[*pipelinesv1.Pipeline]{
		DefinitionCreator: PipelineDefinitionCreator{
			Config: config,
		}.pipelineDefinitionYaml,
		Config:                config,
		TemplateNameGenerator: CompiledTemplateNameGenerator(config),
	}
}
