package pipelines

import (
	"context"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ExperimentWorkflowConstants = struct {
	CreateOperationLabel            string
	DeleteOperationLabel            string
	UpdateOperationLabel            string
	ExperimentIdParameterName 		string
	ExperimentNameLabelKey    		string
	OperationLabelKey               string
	CreationStepName                string
	DeletionStepName                string
}{
	CreateOperationLabel:            "create-experiment",
	DeleteOperationLabel:            "delete-experiment",
	UpdateOperationLabel:            "update-experiment",
	ExperimentIdParameterName: 		 "experiment-id",
	ExperimentNameLabelKey:    		 pipelinesv1.GroupVersion.Group + "/experiment",
	OperationLabelKey:               pipelinesv1.GroupVersion.Group + "/operation",
	CreationStepName:                "create",
	DeletionStepName:                "delete",
}

type ExperimentWorkflowFactory struct {
	WorkflowFactory
}

func (workflows *ExperimentWorkflowFactory) commonMeta(ctx context.Context, rc *pipelinesv1.Experiment, operation string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: operation + "-",
		Namespace:    rc.Namespace,
		Labels: map[string]string{
			ExperimentWorkflowConstants.OperationLabelKey:            operation,
			ExperimentWorkflowConstants.ExperimentNameLabelKey: rc.Name,
		},
		Annotations: workflows.Annotations(ctx, rc.ObjectMeta),
	}
}

func (workflows ExperimentWorkflowFactory) ConstructCreationWorkflow(ctx context.Context, experiment *pipelinesv1.Experiment) *argo.Workflow {
	entrypointName := ExperimentWorkflowConstants.CreateOperationLabel

	return &argo.Workflow{
		ObjectMeta: *workflows.commonMeta(ctx, experiment, ExperimentWorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: workflows.Config.Argo.ServiceAccount,
			Entrypoint:         entrypointName,
			Templates: []argo.Template{
				{
					Name: entrypointName,
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     ExperimentWorkflowConstants.CreationStepName,
									Template: ExperimentWorkflowConstants.CreationStepName,
								},
							},
						},
					},
					Outputs: argo.Outputs{
						Parameters: []argo.Parameter{
							{
								Name: ExperimentWorkflowConstants.ExperimentIdParameterName,
								ValueFrom: &argo.ValueFrom{
									Parameter: fmt.Sprintf("{{steps.%s.outputs.result}}",
										ExperimentWorkflowConstants.CreationStepName),
								},
							},
						},
					},
				},
				workflows.creator(experiment),
			},
		},
	}
}

func (workflows *ExperimentWorkflowFactory) ConstructDeletionWorkflow(ctx context.Context, experiment *pipelinesv1.Experiment) *argo.Workflow {
	entrypointName := ExperimentWorkflowConstants.DeleteOperationLabel

	return &argo.Workflow{
		ObjectMeta: *workflows.commonMeta(ctx, experiment, ExperimentWorkflowConstants.DeleteOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: workflows.Config.Argo.ServiceAccount,
			Entrypoint:         entrypointName,
			Templates: []argo.Template{
				{
					Name: entrypointName,
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     ExperimentWorkflowConstants.DeletionStepName,
									Template: ExperimentWorkflowConstants.DeletionStepName,
								},
							},
						},
					},
				},
				workflows.deleter(experiment),
			},
		},
	}
}

func (workflows *ExperimentWorkflowFactory) ConstructUpdateWorkflow(ctx context.Context, experiment *pipelinesv1.Experiment) *argo.Workflow {
	entrypointName := ExperimentWorkflowConstants.UpdateOperationLabel

	return &argo.Workflow{
		ObjectMeta: *workflows.commonMeta(ctx, experiment, ExperimentWorkflowConstants.UpdateOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: workflows.Config.Argo.ServiceAccount,
			Entrypoint:         entrypointName,
			Templates: []argo.Template{
				{
					Name: entrypointName,
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     ExperimentWorkflowConstants.DeletionStepName,
									Template: ExperimentWorkflowConstants.DeletionStepName,
								},
							},
						},
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     ExperimentWorkflowConstants.CreationStepName,
									Template: ExperimentWorkflowConstants.CreationStepName,
									ContinueOn: &argo.ContinueOn{
										Failed: true,
									},
								},
							},
						},
					},
					Outputs: argo.Outputs{
						Parameters: []argo.Parameter{
							{
								Name: ExperimentWorkflowConstants.ExperimentIdParameterName,
								ValueFrom: &argo.ValueFrom{
									Parameter: fmt.Sprintf("{{steps.%s.outputs.result}}",
										ExperimentWorkflowConstants.CreationStepName),
								},
							},
						},
					},
				},
				workflows.deleter(experiment),
				workflows.creator(experiment),
			},
		},
	}
}

func (workflows *ExperimentWorkflowFactory) creator(experiment *pipelinesv1.Experiment) argo.Template {
	kfpScript := workflows.KfpExt(fmt.Sprintf(`experiment create %s | jq -r '."ID"'`,
		experiment.Name))

	return argo.Template{
		Name:     ExperimentWorkflowConstants.CreationStepName,
		Metadata: workflows.Config.Argo.MetadataDefaults,
		Script:   workflows.ScriptTemplate(kfpScript),
	}
}

func (workflows *ExperimentWorkflowFactory) deleter(experiment *pipelinesv1.Experiment) argo.Template {
	kfpScript := workflows.KfpExt(fmt.Sprintf("experiment delete %s", experiment.Status.KfpId))

	return argo.Template{
		Name:     ExperimentWorkflowConstants.DeletionStepName,
		Metadata: workflows.Config.Argo.MetadataDefaults,
		Script:   workflows.ScriptTemplate(fmt.Sprintf("echo y | %s", kfpScript)),
	}
}
