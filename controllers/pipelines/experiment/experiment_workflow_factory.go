package experiment

import (
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	controller "github.com/sky-uk/kfp-operator/controllers/pipelines"
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

func ExperimentWorkflowFactory(config config.KfpControllerConfigSpec) *controller.ResourceWorkflowFactory[*pipelinesv1.Experiment, providers.ExperimentDefinition] {
	return &controller.ResourceWorkflowFactory[*pipelinesv1.Experiment, providers.ExperimentDefinition]{
		DefinitionCreator: ExperimentDefinitionCreator{
			Config: config,
		}.experimentDefinition,
		Config:                config,
		TemplateNameGenerator: controller.SimpleTemplateNameGenerator(config),
	}
}
