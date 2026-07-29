package workflowfactory

import (
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/internal/config"
	"github.com/sky-uk/kfp-operator/pkg/common"
	providers "github.com/sky-uk/kfp-operator/pkg/providers/base"
)

type experimentDefinitionBuilder struct{}

func (b experimentDefinitionBuilder) build(
	experiment *pipelineshub.Experiment,
) (providers.ExperimentDefinition, error) {
	return providers.ExperimentDefinition{
		Name: common.NamespacedName{
			Namespace: experiment.ObjectMeta.Namespace,
			Name:      experiment.ObjectMeta.Name,
		},
		Version:     experiment.ComputeVersion(),
		Description: experiment.Spec.Description,
	}, nil
}

func ExperimentWorkflowFactory(
	config config.ConfigSpec,
) WorkflowFactory[*pipelineshub.Experiment] {
	return simpleWorkflowFactory[*pipelineshub.Experiment, providers.ExperimentDefinition]{
		assembler: workflowAssembler{config: config},
		builder:   experimentDefinitionBuilder{},
	}
}
