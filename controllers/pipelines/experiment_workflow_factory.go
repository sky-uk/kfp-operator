package pipelines

import (
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
)

type ExperimentDefinitionCreator struct {
	Config config.KfpControllerConfigSpec
}

func (edc ExperimentDefinitionCreator) experimentDefinition(experiment *pipelinesv1.Experiment) (providers.ExperimentDefinition, error) {
	return providers.ExperimentDefinition{
		Name:        common.NamespacedName{Namespace: experiment.ObjectMeta.Namespace, Name: experiment.ObjectMeta.Name},
		Version:     experiment.ComputeVersion(),
		Description: experiment.Spec.Description,
	}, nil
}

func ExperimentWorkflowFactory(config config.KfpControllerConfigSpec) *ResourceWorkflowFactory[*pipelinesv1.Experiment, providers.ExperimentDefinition] {
	return &ResourceWorkflowFactory[*pipelinesv1.Experiment, providers.ExperimentDefinition]{
		DefinitionCreator: ExperimentDefinitionCreator{
			Config: config,
		}.experimentDefinition,
		Config:                config,
		TemplateNameGenerator: SimpleTemplateNameGenerator(config),
	}
}
