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

var WorkflowFactoryConstants = struct {
	pipelineIdParameterName   string
	pipelineYamlParameterName string
	compileStepName           string
	uploadStepName            string
	deletionStepName          string
	updateStepName            string
	pipelineYamlFilePath      string
	pipelineIdFilePath        string
}{
	pipelineIdParameterName:   "pipeline-id",
	pipelineYamlParameterName: "pipeline",
	compileStepName:           "compile",
	uploadStepName:            "upload",
	deletionStepName:          "delete",
	updateStepName:            "update",
	pipelineYamlFilePath:      "/tmp/pipeline.yaml",
	pipelineIdFilePath:        "/tmp/pipeline.txt",
}

var (
	// Needs to be passed by reference
	trueValue = true
)

type WorkflowFactory struct {
	Config configv1.Configuration
}

type CompilerConfig struct {
	RootLocation    string            `yaml:"rootLocation"`
	ServingLocation string            `yaml:"servingLocation"`
	Name            string            `yaml:"name"`
	Image           string            `yaml:"image"`
	TfxComponents   string            `yaml:"tfxComponents"`
	Env             map[string]string `yaml:"env"`
	BeamArgs        map[string]string `yaml:"beamArgs"`
}

func (config CompilerConfig) AsYaml() (string, error) {
	configYaml, err := yaml.Marshal(&config)

	if err != nil {
		return "", err
	}

	return string(configYaml), nil
}

// TODO: Join paths properly (path.Join or filepath.Join don't work with URLs)
func (wf *WorkflowFactory) newCompilerConfig(pipelineSpec pipelinesv1.PipelineSpec, pipelineMeta metav1.ObjectMeta) *CompilerConfig {
	// TODO: should come from config
	servingPath := "/serving"
	tempPath := "/tmp"

	pipelineRoot := wf.Config.PipelineStorage + "/" + pipelineMeta.Name

	beamArgs := make(map[string]string)
	for key, value := range wf.Config.DefaultBeamArgs {
		beamArgs[key] = value
	}
	for key, value := range pipelineSpec.BeamArgs {
		beamArgs[key] = value
	}
	beamArgs["temp_location"] = pipelineRoot + tempPath

	return &CompilerConfig{
		RootLocation:    pipelineRoot,
		ServingLocation: pipelineRoot + servingPath,
		Name:            pipelineMeta.Name,
		Image:           pipelineSpec.Image,
		TfxComponents:   pipelineSpec.TfxComponents,
		Env:             pipelineSpec.Env,
		BeamArgs:        beamArgs,
	}
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
	compilerConfigYaml, error := w.newCompilerConfig(pipelineSpec, pipelineMeta).AsYaml()

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
									Name:     WorkflowFactoryConstants.compileStepName,
									Template: WorkflowFactoryConstants.compileStepName,
								},
							},
						},
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     WorkflowFactoryConstants.uploadStepName,
									Template: WorkflowFactoryConstants.uploadStepName,
									Arguments: argo.Arguments{
										Artifacts: []argo.Artifact{
											{
												Name: WorkflowFactoryConstants.pipelineYamlParameterName,
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
									Name:     WorkflowFactoryConstants.updateStepName,
									Template: WorkflowFactoryConstants.updateStepName,
									Arguments: argo.Arguments{
										Artifacts: []argo.Artifact{
											{
												Name: WorkflowFactoryConstants.pipelineYamlParameterName,
												From: "{{steps.compile.outputs.artifacts.pipeline}}",
											},
										},
										Parameters: []argo.Parameter{
											{
												Name:  WorkflowFactoryConstants.pipelineIdParameterName,
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
								Name: WorkflowFactoryConstants.pipelineIdParameterName,
								ValueFrom: &argo.ValueFrom{
									Parameter: "{{steps.upload.outputs.result}}",
								},
							},
						},
					},
				},
				w.compiler(compilerConfigYaml, pipelineSpec.Image),
				w.uploader(pipelineMeta.Name),
				w.updater(pipelineVersion),
			},
		},
	}

	return workflow, nil
}

func (w WorkflowFactory) ConstructUpdateWorkflow(pipelineSpec pipelinesv1.PipelineSpec, pipelineMeta metav1.ObjectMeta, pipelineId string, pipelineVersion string) (*argo.Workflow, error) {
	compilerConfigYaml, error := w.newCompilerConfig(pipelineSpec, pipelineMeta).AsYaml()

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
									Name:     WorkflowFactoryConstants.compileStepName,
									Template: WorkflowFactoryConstants.compileStepName,
								},
							},
						},
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     WorkflowFactoryConstants.updateStepName,
									Template: WorkflowFactoryConstants.updateStepName,
									Arguments: argo.Arguments{
										Artifacts: []argo.Artifact{
											{
												Name: WorkflowFactoryConstants.pipelineYamlParameterName,
												From: "{{steps.compile.outputs.artifacts.pipeline}}",
											},
										},
										Parameters: []argo.Parameter{
											{
												Name:  WorkflowFactoryConstants.pipelineIdParameterName,
												Value: argo.AnyStringPtr(pipelineId),
											},
										},
									},
								},
							},
						},
					},
				},
				w.compiler(compilerConfigYaml, pipelineSpec.Image),
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
									Name:     WorkflowFactoryConstants.deletionStepName,
									Template: WorkflowFactoryConstants.deletionStepName,
									Arguments: argo.Arguments{
										Parameters: []argo.Parameter{
											{
												Name:  WorkflowFactoryConstants.pipelineIdParameterName,
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

func (workflows *WorkflowFactory) compiler(compilerConfigYaml string, pipelineImage string) argo.Template {
	compilerVolumeName := "compiler"
	compilerVolumePath := "/compiler"

	args := []string{
		"/compiler/compiler.py",
		"--output_file",
		WorkflowFactoryConstants.pipelineYamlFilePath,
		"--pipeline_config",
		compilerConfigYaml,
	}

	return argo.Template{
		Name: WorkflowFactoryConstants.compileStepName,
		Outputs: argo.Outputs{
			Artifacts: []argo.Artifact{
				{
					Name: WorkflowFactoryConstants.pipelineYamlParameterName,
					Path: WorkflowFactoryConstants.pipelineYamlFilePath,
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
					Name:            WorkflowFactoryConstants.compileStepName,
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
			fmt.Sprintf("kfp --endpoint %s --output json pipeline upload -p %s %s  | jq -r '.\"Pipeline Details\".\"ID\"'",
				workflows.Config.KfpEndpoint, pipelineName, WorkflowFactoryConstants.pipelineYamlFilePath)

	return argo.Template{
		Name: WorkflowFactoryConstants.uploadStepName,
		Inputs: argo.Inputs{
			Artifacts: []argo.Artifact{
				{
					Name: WorkflowFactoryConstants.pipelineYamlParameterName,
					Path: WorkflowFactoryConstants.pipelineYamlFilePath,
				},
			},
		},
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

func (workflows *WorkflowFactory) deleter() argo.Template {
	script :=
		"set -e -o pipefail\n" +
			fmt.Sprintf("kfp --endpoint %s pipeline delete {{inputs.parameters.pipeline-id}}",
				workflows.Config.KfpEndpoint)

	return argo.Template{
		Name: WorkflowFactoryConstants.deletionStepName,
		Inputs: argo.Inputs{
			Parameters: []argo.Parameter{
				{
					Name: WorkflowFactoryConstants.pipelineIdParameterName,
				},
			},
		},
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

func (workflows *WorkflowFactory) updater(version string) argo.Template {
	script :=
		"set -e -o pipefail\n" +
			fmt.Sprintf("kfp --endpoint %s pipeline upload-version -v %s -p {{inputs.parameters.pipeline-id}} %s",
				workflows.Config.KfpEndpoint, version, WorkflowFactoryConstants.pipelineYamlFilePath)

	return argo.Template{
		Name: WorkflowFactoryConstants.updateStepName,
		Inputs: argo.Inputs{
			Artifacts: []argo.Artifact{
				{
					Name: WorkflowFactoryConstants.pipelineYamlParameterName,
					Path: WorkflowFactoryConstants.pipelineYamlFilePath,
				},
			},
			Parameters: []argo.Parameter{
				{
					Name: WorkflowFactoryConstants.pipelineIdParameterName,
				},
			},
		},
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
