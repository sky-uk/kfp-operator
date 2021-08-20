package controllers

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pipelinesv1 "github.com/sky-uk/kfp-operator/api/v1"
)

var constructUploadWorkflow = func(pipeline *pipelinesv1.Pipeline) *argo.Workflow {

	workflow := &argo.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "create-pipeline-" + pipeline.ObjectMeta.Name,
			Namespace: "default",
			Labels: map[string]string{
				"operation": "create-pipeline",
			},
		},
		Spec: argo.WorkflowSpec{
			Entrypoint: "create-pipeline",
			Templates: []argo.Template{
				argo.Template{
					Name: "create-pipeline",
					Steps: []argo.ParallelSteps{
						argo.ParallelSteps{
							Steps: []argo.WorkflowStep{
								argo.WorkflowStep{
									Name:     "uploader",
									Template: "uploader",
								},
							},
						},
					},
					Outputs: argo.Outputs{
						Parameters: []argo.Parameter{
							argo.Parameter{
								Name: "id",
								ValueFrom: &argo.ValueFrom{
									Parameter: "{{steps.uploader.outputs.result}}",
								},
							},
						},
					},
				},
				argo.Template{
					Name: "uploader",
					Script: &argo.ScriptTemplate{
						Container: apiv1.Container{
							Image:   "alpine",
							Command: []string{"sh", "-c"},
						},
						Source: "echo 12345",
					},
				},
			},
		},
	}

	return workflow
}
