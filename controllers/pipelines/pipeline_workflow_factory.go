package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha2"
	"gopkg.in/yaml.v2"
)

var PipelineWorkflowConstants = struct {
	PipelineIdParameterName      string
	PipelineNameParameterName    string
	PipelineImageParameterName   string
	PipelineVersionParameterName string
	CompilerConfigParameterName  string
}{
	PipelineIdParameterName:      "pipeline-id",
	PipelineNameParameterName:    "pipeline-name",
	PipelineImageParameterName:   "pipeline-image",
	PipelineVersionParameterName: "pipeline-version",
	CompilerConfigParameterName:  "compiler-config",
}

type PipelineWorkflowFactory struct {
	WorkflowFactory
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
func (wf *PipelineWorkflowFactory) newCompilerConfig(pipeline *pipelinesv1.Pipeline) *CompilerConfig {
	// TODO: should come from config
	servingPath := "/serving"
	tempPath := "/tmp"

	pipelineRoot := wf.Config.PipelineStorage + "/" + pipeline.ObjectMeta.Name

	beamArgs := make(map[string]string)
	for key, value := range wf.Config.DefaultBeamArgs {
		beamArgs[key] = value
	}
	for key, value := range pipeline.Spec.BeamArgs {
		beamArgs[key] = value
	}
	beamArgs["temp_location"] = pipelineRoot + tempPath

	return &CompilerConfig{
		RootLocation:    pipelineRoot,
		ServingLocation: pipelineRoot + servingPath,
		Name:            pipeline.ObjectMeta.Name,
		Image:           pipeline.Spec.Image,
		TfxComponents:   pipeline.Spec.TfxComponents,
		Env:             pipeline.Spec.Env,
		BeamArgs:        beamArgs,
	}
}

func (w PipelineWorkflowFactory) ConstructCreationWorkflow(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	compilerConfigYaml, err := w.newCompilerConfig(pipeline).AsYaml()

	if err != nil {
		return nil, err
	}

	workflow := &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(pipeline, WorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name: PipelineWorkflowConstants.CompilerConfigParameterName,
						Value:  argo.AnyStringPtr(compilerConfigYaml),
					},
					{
						Name: PipelineWorkflowConstants.PipelineImageParameterName,
						Value:  argo.AnyStringPtr(pipeline.Spec.Image),
					},
					{
						Name: PipelineWorkflowConstants.PipelineNameParameterName,
						Value:  argo.AnyStringPtr(pipeline.Name),
					},
					{
						Name: PipelineWorkflowConstants.PipelineVersionParameterName,
						Value:  argo.AnyStringPtr(pipeline.Spec.ComputeVersion()),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
						Name:         "create-pipeline",
						ClusterScope: true,
			},
		},
	}

	return workflow, nil
}

func (w PipelineWorkflowFactory) ConstructUpdateWorkflow(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	compilerConfigYaml, err := w.newCompilerConfig(pipeline).AsYaml()

	if err != nil {
		return nil, err
	}

	workflow := &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(pipeline, WorkflowConstants.UpdateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name: PipelineWorkflowConstants.CompilerConfigParameterName,
						Value:  argo.AnyStringPtr(compilerConfigYaml),
					},
					{
						Name: PipelineWorkflowConstants.PipelineIdParameterName,
						Value:  argo.AnyStringPtr(pipeline.Status.KfpId),
					},
					{
						Name: PipelineWorkflowConstants.PipelineImageParameterName,
						Value:  argo.AnyStringPtr(pipeline.Spec.Image),
					},
					{
						Name: PipelineWorkflowConstants.PipelineVersionParameterName,
						Value:  argo.AnyStringPtr(pipeline.Spec.ComputeVersion()),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         "update-pipeline",
				ClusterScope: true,
			},
		},
	}

	return workflow, nil
}

func (w PipelineWorkflowFactory) ConstructDeletionWorkflow(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(pipeline, WorkflowConstants.DeleteOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name: PipelineWorkflowConstants.PipelineIdParameterName,
						Value:  argo.AnyStringPtr(pipeline.Status.KfpId),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         "delete-pipeline",
				ClusterScope: true,
			},
		},
	}, nil
}
