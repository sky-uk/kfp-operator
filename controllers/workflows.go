package controllers

import (
	"strings"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pipelinesv1 "github.com/sky-uk/kfp-operator/api/v1"
)

var printMap = func(m map[string]string) string {
	outputs := []string{}

	for k, v := range m {
		outputs = append(outputs, k+"="+v)
	}

	return strings.Join(outputs, ",")
}

var constructUploadWorkflow = func(pipeline *pipelinesv1.Pipeline) *argo.Workflow {
	image := argo.AnyString(pipeline.Spec.Image)
	tfxComponents := argo.AnyString(pipeline.Spec.TfxComponents)
	env := argo.AnyString(printMap(pipeline.Spec.Env))

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
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					argo.Parameter{
						Name:  "image",
						Value: &image,
					},
					argo.Parameter{
						Name:  "tfxComponents",
						Value: &tfxComponents,
					},
					argo.Parameter{
						Name:  "tfxComponents",
						Value: &tfxComponents,
					},
					argo.Parameter{
						Name:  "env",
						Value: &env,
					},
				},
			},
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
