package pipelines

import (
	"context"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var RunConfigurationWorkflowConstants = struct {
	CreateOperationLabel            string
	DeleteOperationLabel            string
	UpdateOperationLabel            string
	RunConfigurationIdParameterName string
	RunConfigurationNameLabelKey    string
	OperationLabelKey               string
	CreationStepName                string
	DeletionStepName                string
}{
	CreateOperationLabel:            "create-runconfiguration",
	DeleteOperationLabel:            "delete-runconfiguration",
	UpdateOperationLabel:            "update-runconfiguration",
	RunConfigurationIdParameterName: "runconfiguration-id",
	RunConfigurationNameLabelKey:    pipelinesv1.GroupVersion.Group + "/runConfiguration",
	OperationLabelKey:               pipelinesv1.GroupVersion.Group + "/operation",
	CreationStepName:                "create",
	DeletionStepName:                "delete",
}

type RunConfigurationWorkflowFactory struct {
	WorkflowFactory
}

// TODO: use input parameters

func (workflows *RunConfigurationWorkflowFactory) commonMeta(ctx context.Context, rc *pipelinesv1.RunConfiguration, operation string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: operation + "-",
		Namespace:    rc.Namespace,
		Labels: map[string]string{
			RunConfigurationWorkflowConstants.OperationLabelKey:            operation,
			RunConfigurationWorkflowConstants.RunConfigurationNameLabelKey: rc.Name,
		},
		Annotations: workflows.Annotations(ctx, rc.ObjectMeta),
	}
}

func (workflows RunConfigurationWorkflowFactory) ConstructCreationWorkflow(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) *argo.Workflow {
	entrypointName := RunConfigurationWorkflowConstants.CreateOperationLabel

	return &argo.Workflow{
		ObjectMeta: *workflows.commonMeta(ctx, runConfiguration, RunConfigurationWorkflowConstants.CreateOperationLabel),
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
									Name:     RunConfigurationWorkflowConstants.CreationStepName,
									Template: RunConfigurationWorkflowConstants.CreationStepName,
								},
							},
						},
					},
					Outputs: argo.Outputs{
						Parameters: []argo.Parameter{
							{
								Name: RunConfigurationWorkflowConstants.RunConfigurationIdParameterName,
								ValueFrom: &argo.ValueFrom{
									Parameter: fmt.Sprintf("{{steps.%s.outputs.result}}",
										RunConfigurationWorkflowConstants.CreationStepName),
								},
							},
						},
					},
				},
				workflows.creator(runConfiguration),
			},
		},
	}
}

func (workflows *RunConfigurationWorkflowFactory) ConstructDeletionWorkflow(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) *argo.Workflow {
	entrypointName := RunConfigurationWorkflowConstants.DeleteOperationLabel

	return &argo.Workflow{
		ObjectMeta: *workflows.commonMeta(ctx, runConfiguration, RunConfigurationWorkflowConstants.DeleteOperationLabel),
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
									Name:     RunConfigurationWorkflowConstants.DeletionStepName,
									Template: RunConfigurationWorkflowConstants.DeletionStepName,
								},
							},
						},
					},
				},
				workflows.deleter(runConfiguration),
			},
		},
	}
}

func (workflows *RunConfigurationWorkflowFactory) ConstructUpdateWorkflow(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) *argo.Workflow {
	entrypointName := RunConfigurationWorkflowConstants.UpdateOperationLabel

	return &argo.Workflow{
		ObjectMeta: *workflows.commonMeta(ctx, runConfiguration, RunConfigurationWorkflowConstants.UpdateOperationLabel),
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
									Name:     RunConfigurationWorkflowConstants.DeletionStepName,
									Template: RunConfigurationWorkflowConstants.DeletionStepName,
								},
							},
						},
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     RunConfigurationWorkflowConstants.CreationStepName,
									Template: RunConfigurationWorkflowConstants.CreationStepName,
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
								Name: RunConfigurationWorkflowConstants.RunConfigurationIdParameterName,
								ValueFrom: &argo.ValueFrom{
									Parameter: fmt.Sprintf("{{steps.%s.outputs.result}}",
										RunConfigurationWorkflowConstants.CreationStepName),
								},
							},
						},
					},
				},
				workflows.deleter(runConfiguration),
				workflows.creator(runConfiguration),
			},
		},
	}
}

func (workflows *RunConfigurationWorkflowFactory) creator(runConfiguration *pipelinesv1.RunConfiguration) argo.Template {
	kfpScript := workflows.KfpExt(fmt.Sprintf(`job submit --experiment-name %s --job-name %s --pipeline-name %s --cron-expression '%s' | jq -r '."Job Details"."ID"'`,
		workflows.Config.DefaultExperiment, runConfiguration.Name, runConfiguration.Spec.PipelineName, runConfiguration.Spec.Schedule))

	return argo.Template{
		Name:     RunConfigurationWorkflowConstants.CreationStepName,
		Metadata: workflows.Config.Argo.MetadataDefaults,
		Script:   workflows.ScriptTemplate(kfpScript),
	}
}

func (workflows *RunConfigurationWorkflowFactory) deleter(runConfiguration *pipelinesv1.RunConfiguration) argo.Template {
	kfpScript := workflows.KfpExt(fmt.Sprintf("job delete %s", runConfiguration.Status.KfpId))

	return argo.Template{
		Name:     RunConfigurationWorkflowConstants.DeletionStepName,
		Metadata: workflows.Config.Argo.MetadataDefaults,
		Script:   workflows.ScriptTemplate(kfpScript),
	}
}
