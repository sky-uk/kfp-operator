package workflowfactory

import (
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/pkg/common"
	providers "github.com/sky-uk/kfp-operator/pkg/providers/base"
)

type ExperimentDefinitionCreator struct {
	Config config.KfpControllerConfigSpec
}

func (edc ExperimentDefinitionCreator) experimentDefinition(
	_ pipelineshub.Provider,
	experiment *pipelineshub.Experiment,
) ([]pipelineshub.Patch, providers.ExperimentDefinition, error) {
	return nil, providers.ExperimentDefinition{
		Name: common.NamespacedName{
			Namespace: experiment.ObjectMeta.Namespace,
			Name:      experiment.ObjectMeta.Name,
		},
		Version:     experiment.ComputeVersion(),
		Description: experiment.Spec.Description,
	}, nil
}

func ExperimentWorkflowFactory(
	config config.KfpControllerConfigSpec,
) *ResourceWorkflowFactory[*pipelineshub.Experiment, providers.ExperimentDefinition] {
	return NewResourceWorkflowFactory(
		config,
		SimpleSuffix,
		ExperimentDefinitionCreator{Config: config}.experimentDefinition,
		WorkflowParamsCreatorNoop[*pipelineshub.Experiment],
	)
}
