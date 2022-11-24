package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	providers "github.com/sky-uk/kfp-operator/providers/base"
	"gopkg.in/yaml.v2"
)

var ExperimentWorkflowConstants = struct {
	ExperimentDefinitionParameterName string
	ExperimentIdParameterName         string
}{
	ExperimentDefinitionParameterName: "experiment-definition",
	ExperimentIdParameterName:         "experiment-id",
}

type ExperimentWorkflowFactory struct {
	WorkflowFactoryBase
}

func (wf *ExperimentWorkflowFactory) experimentDefinitionYaml(experiment *pipelinesv1.Experiment) (string, error) {
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

func (workflows ExperimentWorkflowFactory) ConstructCreationWorkflow(experiment *pipelinesv1.Experiment) (*argo.Workflow, error) {
	experimentDefinition, err := workflows.experimentDefinitionYaml(experiment)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *workflows.CommonWorkflowMeta(experiment, WorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  ExperimentWorkflowConstants.ExperimentDefinitionParameterName,
						Value: argo.AnyStringPtr(experimentDefinition),
					},
					{
						Name:  WorkflowConstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(workflows.ProviderConfig),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.Config.WorkflowTemplatePrefix + "create-experiment",
			},
		},
	}, nil
}

func (workflows ExperimentWorkflowFactory) ConstructUpdateWorkflow(experiment *pipelinesv1.Experiment) (*argo.Workflow, error) {
	experimentDefinition, err := workflows.experimentDefinitionYaml(experiment)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *workflows.CommonWorkflowMeta(experiment, WorkflowConstants.UpdateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  ExperimentWorkflowConstants.ExperimentDefinitionParameterName,
						Value: argo.AnyStringPtr(experimentDefinition),
					},
					{
						Name:  ExperimentWorkflowConstants.ExperimentIdParameterName,
						Value: argo.AnyStringPtr(experiment.Status.ProviderId),
					},
					{
						Name:  WorkflowConstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(workflows.ProviderConfig),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.Config.WorkflowTemplatePrefix + "update-experiment",
			},
		},
	}, nil
}

func (workflows ExperimentWorkflowFactory) ConstructDeletionWorkflow(experiment *pipelinesv1.Experiment) (*argo.Workflow, error) {
	return &argo.Workflow{
		ObjectMeta: *workflows.CommonWorkflowMeta(experiment, WorkflowConstants.DeleteOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  ExperimentWorkflowConstants.ExperimentIdParameterName,
						Value: argo.AnyStringPtr(experiment.Status.ProviderId),
					},
					{
						Name:  WorkflowConstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(workflows.ProviderConfig),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.Config.WorkflowTemplatePrefix + "delete-experiment",
			},
		},
	}, nil
}
