package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha2"
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
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         "create-experiment",
				ClusterScope: true,
			},
		},
	}, nil
}

func (workflows *ExperimentWorkflowFactory) ConstructUpdateWorkflow(experiment *pipelinesv1.Experiment) (*argo.Workflow, error) {
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
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         "update-experiment",
				ClusterScope: true,
			},
		},
	}, nil
}

func (workflows *ExperimentWorkflowFactory) ConstructDeletionWorkflow(experiment *pipelinesv1.Experiment) (*argo.Workflow, error) {
	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(experiment, WorkflowConstants.DeleteOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  ExperimentWorkflowConstants.ExperimentIdParameterName,
						Value: argo.AnyStringPtr(experiment.Status.KfpId),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         "delete-experiment",
				ClusterScope: true,
			},
		},
	}, nil
}
