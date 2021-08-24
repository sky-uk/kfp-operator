package controllers

import (
	"fmt"

	"gopkg.in/yaml.v2"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pipelinesv1 "github.com/sky-uk/kfp-operator/api/v1"
)

var constructUploadWorkflow = func(pipeline *pipelinesv1.Pipeline) *argo.Workflow {
	yml, err := yaml.Marshal(&pipeline.Spec)

	fmt.Print(argo.AnyString(yml))
	if err != nil {
		return nil
	}

	configuration := argo.AnyString(yml)

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
					{
						Name:  "config",
						Value: &configuration,
					},
				},
			},
			Templates: []argo.Template{
				{
					Name: "create-pipeline",
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     "uploader",
									Template: "uploader",
								},
							},
						},
					},
					Outputs: argo.Outputs{
						Parameters: []argo.Parameter{
							{
								Name: "id",
								ValueFrom: &argo.ValueFrom{
									Parameter: "{{steps.uploader.outputs.result}}",
								},
							},
						},
					},
				},
				{
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
