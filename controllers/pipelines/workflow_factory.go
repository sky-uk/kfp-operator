package pipelines

import (
	"fmt"

	"gopkg.in/yaml.v2"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PipelineConfigKey         = "pipeline-config"
	PipelineIdParameterName   = "pipeline-id"
	PipelineYamlParameterName = "pipeline"
	CompileStepName           = "compile"
	UploadStepName            = "upload"
	DeletionStepName          = "delete"
	UpdateStepName            = "update"
	PipelineYamlFilePath      = "/tmp/pipeline.yaml"
	PipelineIdFilePath        = "/tmp/pipeline.txt"
)

var (
	trueValue = true
)

type WorkflowFactory struct {
	Config configv1.Configuration
}

func (wf *WorkflowFactory) pipelineConfigAsYaml(pipelineSpec pipelinesv1.PipelineSpec, pipelineMeta metav1.ObjectMeta) (string, error) {
	type CompilerConfig struct {
		Spec         pipelinesv1.PipelineSpec
		Name         string
		ServingDir   string            `yaml:"servingDir"`
		PipelineRoot string            `yaml:"pipelineRoot"`
		BeamArgs     map[string]string `yaml:"beamArgs"`
	}

	// TODO: Join paths properly
	pipelineRoot := wf.Config.PipelineStorage + "/" + pipelineMeta.Name
	servingDir := pipelineRoot + "/serving"
	dataflowTmpDir := pipelineRoot + "/tmp"

	config := &CompilerConfig{
		Spec:         pipelineSpec,
		Name:         pipelineMeta.Name,
		PipelineRoot: pipelineRoot,
		ServingDir:   servingDir,
		BeamArgs: map[string]string{
			"project":       wf.Config.DataflowProject,
			"temp_location": dataflowTmpDir,
		},
	}

	specAsYaml, err := yaml.Marshal(&config)

	if err != nil {
		return "", err
	}

	return string(specAsYaml), nil
}

func (w *WorkflowFactory) commonMeta(pipelineMeta metav1.ObjectMeta, operation string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: operation + "-",
		Namespace:    pipelineMeta.Namespace,
		Labels: map[string]string{
			OperationLabelKey:    operation,
			PipelineNameLabelKey: pipelineMeta.Name,
		},
	}
}

func (w WorkflowFactory) ConstructCreationWorkflow(pipelineSpec pipelinesv1.PipelineSpec, pipelineMeta metav1.ObjectMeta, pipelineVersion string) (*argo.Workflow, error) {
	yamlConfig, error := w.pipelineConfigAsYaml(pipelineSpec, pipelineMeta)

	if error != nil {
		return nil, error
	}

	entrypointName := CreateOperationLabel

	workflow := &argo.Workflow{
		ObjectMeta: *w.commonMeta(pipelineMeta, CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: w.Config.ServiceAccount,
			Entrypoint:         entrypointName,
			Templates: []argo.Template{
				{
					Name: entrypointName,
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     CompileStepName,
									Template: CompileStepName,
								},
							},
						},
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     UploadStepName,
									Template: UploadStepName,
									Arguments: argo.Arguments{
										Artifacts: []argo.Artifact{
											{
												Name: PipelineYamlParameterName,
												From: "{{steps.compile.outputs.artifacts.pipeline}}",
											},
										},
									},
								},
							},
						},
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     UpdateStepName,
									Template: UpdateStepName,
									Arguments: argo.Arguments{
										Artifacts: []argo.Artifact{
											{
												Name: PipelineYamlParameterName,
												From: "{{steps.compile.outputs.artifacts.pipeline}}",
											},
										},
										Parameters: []argo.Parameter{
											{
												Name:  PipelineIdParameterName,
												Value: argo.AnyStringPtr("{{steps.upload.outputs.result}}"),
											},
										},
									},
								},
							},
						},
					},
					Outputs: argo.Outputs{
						Parameters: []argo.Parameter{
							{
								Name: PipelineIdParameterName,
								ValueFrom: &argo.ValueFrom{
									Parameter: "{{steps.upload.outputs.result}}",
								},
							},
						},
					},
				},
				w.compiler(yamlConfig, pipelineSpec.Image),
				w.uploader(pipelineMeta.Name),
				w.updater(pipelineVersion),
			},
		},
	}

	return workflow, nil
}

func (w WorkflowFactory) ConstructUpdateWorkflow(pipelineSpec pipelinesv1.PipelineSpec, pipelineMeta metav1.ObjectMeta, pipelineId string, pipelineVersion string) (*argo.Workflow, error) {
	yamlConfig, error := w.pipelineConfigAsYaml(pipelineSpec, pipelineMeta)

	if error != nil {
		return nil, error
	}

	entrypointName := UpdateOperationLabel

	workflow := &argo.Workflow{
		ObjectMeta: *w.commonMeta(pipelineMeta, UpdateOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: w.Config.ServiceAccount,
			Entrypoint:         entrypointName,
			Templates: []argo.Template{
				{
					Name: entrypointName,
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     CompileStepName,
									Template: CompileStepName,
								},
							},
						},
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     UpdateStepName,
									Template: UpdateStepName,
									Arguments: argo.Arguments{
										Artifacts: []argo.Artifact{
											{
												Name: PipelineYamlParameterName,
												From: "{{steps.compile.outputs.artifacts.pipeline}}",
											},
										},
										Parameters: []argo.Parameter{
											{
												Name:  PipelineIdParameterName,
												Value: argo.AnyStringPtr(pipelineId),
											},
										},
									},
								},
							},
						},
					},
				},
				w.compiler(yamlConfig, pipelineSpec.Image),
				w.updater(pipelineVersion),
			},
		},
	}

	return workflow, nil
}

func (w WorkflowFactory) ConstructDeletionWorkflow(pipelineMeta metav1.ObjectMeta, pipelineId string) *argo.Workflow {

	entrypointName := DeleteOperationLabel

	workflow := &argo.Workflow{
		ObjectMeta: *w.commonMeta(pipelineMeta, DeleteOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: w.Config.ServiceAccount,
			Entrypoint:         entrypointName,
			Templates: []argo.Template{
				{
					Name: entrypointName,
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     DeletionStepName,
									Template: DeletionStepName,
									Arguments: argo.Arguments{
										Parameters: []argo.Parameter{
											{
												Name:  PipelineIdParameterName,
												Value: argo.AnyStringPtr(pipelineId),
											},
										},
									},
								},
							},
						},
					},
				},
				w.deleter(),
			},
		},
	}

	return workflow
}

func (workflows *WorkflowFactory) compiler(pipelineSpec string, pipelineImage string) argo.Template {
	compilerVolumeName := "compiler"
	compilerVolumePath := "/compiler"

	args := []string{
		"/compiler/compiler.py",
		"--output_file",
		PipelineYamlFilePath,
		"--pipeline_config",
		pipelineSpec,
	}

	return argo.Template{
		Name: CompileStepName,
		Outputs: argo.Outputs{
			Artifacts: []argo.Artifact{
				{
					Name: PipelineYamlParameterName,
					Path: PipelineYamlFilePath,
				},
			},
		},
		Container: &apiv1.Container{
			Name:            "pipeline",
			Image:           pipelineImage,
			ImagePullPolicy: apiv1.PullPolicy(workflows.Config.ImagePullPolicy),
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      compilerVolumeName,
					MountPath: compilerVolumePath,
				},
			},
			Command: []string{"python3"},
			Args:    args,
		},
		InitContainers: []argo.UserContainer{
			{
				Container: apiv1.Container{
					Name:            CompileStepName,
					Image:           workflows.Config.CompilerImage,
					ImagePullPolicy: apiv1.PullPolicy(workflows.Config.ImagePullPolicy),
				},
				MirrorVolumeMounts: &trueValue,
			},
		},
		Volumes: []apiv1.Volume{
			{
				Name: compilerVolumeName,
			},
		},
	}
}

func (workflows *WorkflowFactory) uploader(pipelineName string) argo.Template {
	script :=
		"set -e -o pipefail\n" +
			fmt.Sprintf("kfp --endpoint %s --output json pipeline upload -p %s %s  | jq -r '.\"Pipeline Details\".\"ID\"'", workflows.Config.KfpEndpoint, pipelineName, PipelineYamlFilePath)

	return argo.Template{
		Name: UploadStepName,
		Inputs: argo.Inputs{
			Artifacts: []argo.Artifact{
				{
					Name: PipelineYamlParameterName,
					Path: PipelineYamlFilePath,
				},
			},
		},
		Script: &argo.ScriptTemplate{
			Container: apiv1.Container{
				Image:           workflows.Config.KfpToolsImage,
				ImagePullPolicy: apiv1.PullPolicy(workflows.Config.ImagePullPolicy),
				Command:         []string{"ash"},
			},
			Source: script,
		},
	}
}

func (workflows *WorkflowFactory) deleter() argo.Template {
	script :=
		"set -e -o pipefail\n" +
			fmt.Sprintf("kfp --endpoint %s pipeline delete {{inputs.parameters.pipeline-id}}", workflows.Config.KfpEndpoint)

	return argo.Template{
		Name: DeletionStepName,
		Inputs: argo.Inputs{
			Parameters: []argo.Parameter{
				{
					Name: PipelineIdParameterName,
				},
			},
		},
		Script: &argo.ScriptTemplate{
			Container: apiv1.Container{
				Image:           workflows.Config.KfpToolsImage,
				ImagePullPolicy: apiv1.PullPolicy(workflows.Config.ImagePullPolicy),
				Command:         []string{"ash"},
			},
			Source: script,
		},
	}
}

func (workflows *WorkflowFactory) updater(version string) argo.Template {
	script :=
		"set -e -o pipefail\n" +
			fmt.Sprintf("kfp --endpoint %s pipeline upload-version -v %s -p {{inputs.parameters.pipeline-id}} %s", workflows.Config.KfpEndpoint, version, PipelineYamlFilePath)

	return argo.Template{
		Name: UpdateStepName,
		Inputs: argo.Inputs{
			Artifacts: []argo.Artifact{
				{
					Name: PipelineYamlParameterName,
					Path: PipelineYamlFilePath,
				},
			},
			Parameters: []argo.Parameter{
				{
					Name: PipelineIdParameterName,
				},
			},
		},
		Script: &argo.ScriptTemplate{
			Container: apiv1.Container{
				Image:           workflows.Config.KfpToolsImage,
				ImagePullPolicy: apiv1.PullPolicy(workflows.Config.ImagePullPolicy),
				Command:         []string{"ash"},
			},
			Source: script,
		},
	}
}
