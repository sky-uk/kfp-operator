package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
)

var ExperimentWorkflowConstants = struct {
	ExperimentIdParameterName   string
	ExperimentNameParameterName string
}{
	ExperimentIdParameterName:   "experiment-id",
	ExperimentNameParameterName: "experiment-name",
}

type ExperimentWorkflowFactory struct {
	WorkflowFactoryBase
}

func (workflows ExperimentWorkflowFactory) ConstructCreationWorkflow(experiment *pipelinesv1.Experiment) (*argo.Workflow, error) {
	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(experiment, WorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  ExperimentWorkflowConstants.ExperimentNameParameterName,
						Value: argo.AnyStringPtr(experiment.Name),
					},
					{
						Name:  WorkflowConstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(workflows.ProviderConfig),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         workflows.Config.WorkflowTemplatePrefix + "create-experiment",
				ClusterScope: true,
			},
		},
	}, nil
}

func (workflows ExperimentWorkflowFactory) ConstructUpdateWorkflow(experiment *pipelinesv1.Experiment) (*argo.Workflow, error) {
	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(experiment, WorkflowConstants.UpdateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  ExperimentWorkflowConstants.ExperimentIdParameterName,
						Value: argo.AnyStringPtr(experiment.Status.KfpId),
					},
					{
						Name:  ExperimentWorkflowConstants.ExperimentNameParameterName,
						Value: argo.AnyStringPtr(experiment.Name),
					},
					{
						Name:  WorkflowConstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(workflows.ProviderConfig),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         workflows.Config.WorkflowTemplatePrefix + "update-experiment",
				ClusterScope: true,
			},
		},
	}, nil
}

func (workflows ExperimentWorkflowFactory) ConstructDeletionWorkflow(experiment *pipelinesv1.Experiment) (*argo.Workflow, error) {
	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(experiment, WorkflowConstants.DeleteOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  ExperimentWorkflowConstants.ExperimentIdParameterName,
						Value: argo.AnyStringPtr(experiment.Status.KfpId),
					},
					{
						Name:  WorkflowConstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(workflows.ProviderConfig),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         workflows.Config.WorkflowTemplatePrefix + "delete-experiment",
				ClusterScope: true,
			},
		},
	}, nil
}
