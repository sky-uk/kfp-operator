package pipelines

import (
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	providers "github.com/sky-uk/kfp-operator/providers/base"
	"gopkg.in/yaml.v2"
)

type ExperimentDefinitionCreator struct {
	Config config.Configuration
}

func (edc ExperimentDefinitionCreator) experimentDefinitionYaml(experiment *pipelinesv1.Experiment) (string, error) {
	experimentDefinition := providers.ExperimentDefinition{
		Name:        experiment.ObjectMeta.Name,
		Version:     experiment.ComputeVersion(),
		Description: experiment.Spec.Description,
	}

	marshalled, err := yaml.Marshal(&experimentDefinition)
	if err != nil {
		return "", err
	}

	return string(marshalled), nil
}

func ExperimentWorkflowFactory(config config.Configuration) ResourceWorkflowFactory[*pipelinesv1.Experiment] {
	return ResourceWorkflowFactory[*pipelinesv1.Experiment]{
		DefinitionCreator: ExperimentDefinitionCreator{
			Config: config,
		}.experimentDefinitionYaml,
		Config: config,
	}
}
