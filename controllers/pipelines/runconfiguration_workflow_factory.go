package pipelines

import (
	"fmt"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	providers "github.com/sky-uk/kfp-operator/providers/base"
	"gopkg.in/yaml.v2"
)

type RunConfigurationDefinitionCreator struct {
	Config config.Configuration
}

func (rcdc RunConfigurationDefinitionCreator) runConfigurationDefinitionYaml(runConfiguration *pipelinesv1.RunConfiguration) (string, error) {
	var experimentName string

	if runConfiguration.Spec.ExperimentName == "" {
		experimentName = rcdc.Config.DefaultExperiment
	} else {
		experimentName = runConfiguration.Spec.ExperimentName
	}

	if runConfiguration.Status.ObservedPipelineVersion == "" {
		return "", fmt.Errorf("unknown pipeline version")
	}

	runConfigurationDefinition := providers.RunConfigurationDefinition{
		Name:            runConfiguration.ObjectMeta.Name,
		Version:         runConfiguration.ComputeVersion(),
		PipelineName:    runConfiguration.Spec.Pipeline.Name,
		PipelineVersion: runConfiguration.Status.ObservedPipelineVersion,
		ExperimentName:  experimentName,
		Schedule:        runConfiguration.Spec.Schedule,
	}

	marshalled, err := yaml.Marshal(&runConfigurationDefinition)

	if err != nil {
		return "", err
	}

	return string(marshalled), nil
}

func RunConfigurationWorkflowFactory(config config.Configuration) ResourceWorkflowFactory[*pipelinesv1.RunConfiguration] {
	return ResourceWorkflowFactory[*pipelinesv1.RunConfiguration]{
		DefinitionCreator: RunConfigurationDefinitionCreator{
			Config: config,
		}.runConfigurationDefinitionYaml,
		Config: config,
	}
}
