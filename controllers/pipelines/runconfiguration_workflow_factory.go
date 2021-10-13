package pipelines

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var RunConfigurationWorkflowConstants = struct {
	CreateOperationLabel 	  string
	RunConfigurationIdParameterName   string
	RunConfigurationNameLabelKey string
	OperationLabelKey string
	CreationStepName string
}{
	CreateOperationLabel:      "create-runconfiguration",
	RunConfigurationIdParameterName:   "runconfiguration-id",
	RunConfigurationNameLabelKey:      "pipelines.kubeflow.org/runConfiguration",
	OperationLabelKey:         "pipelines.kubeflow.org/operation",
	CreationStepName: "create",
}

type RunConfigurationWorkflowFactory struct {
	Config configv1.Configuration
}

func (w *RunConfigurationWorkflowFactory) commonMeta(rc *pipelinesv1.RunConfiguration, operation string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: operation + "-",
		Namespace:    rc.Namespace,
		Labels: map[string]string{
			RunConfigurationWorkflowConstants.OperationLabelKey:    operation,
			RunConfigurationWorkflowConstants.RunConfigurationNameLabelKey: rc.Name,
		},
	}
}

func (w RunConfigurationWorkflowFactory) ConstructCreationWorkflow(runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	entrypointName := RunConfigurationWorkflowConstants.CreateOperationLabel

	workflow := argo.Workflow{
		ObjectMeta: *w.commonMeta(runConfiguration, RunConfigurationWorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: w.Config.ServiceAccount,
			Entrypoint: entrypointName,
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
									Parameter: "{{steps.create.outputs.result}}",
								},
							},
						},
					},
				},
				w.creator(runConfiguration),
			},
		},
	}

	return &workflow, nil
}

func (workflows *RunConfigurationWorkflowFactory) creator(runConfiguration *pipelinesv1.RunConfiguration) argo.Template {
	script :=
		"set -e -o pipefail\n" +
			fmt.Sprintf("kfp-ext --endpoint %s --output json job submit --experiment-name %s --job-name %s --pipeline-name %s --cron-expression '%s' | jq -r '.\"Job Details\".\"ID\"'",
				workflows.Config.KfpEndpoint, workflows.Config.DefaultExperiment, runConfiguration.Name, runConfiguration.Spec.PipelineName, runConfiguration.Spec.Schedule)

	return argo.Template{
		Name: RunConfigurationWorkflowConstants.CreationStepName,
		Script: &argo.ScriptTemplate{
			Container: apiv1.Container{
				Image:           workflows.Config.KfpSdkImage,
				ImagePullPolicy: apiv1.PullPolicy(workflows.Config.ImagePullPolicy),
				Command:         []string{"ash"},
			},
			Source: script,
		},
	}
}
