package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
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
	WorkflowFactoryBase
}

type CompilerConfig struct {
	RootLocation    string              `yaml:"rootLocation"`
	ServingLocation string              `yaml:"servingLocation"`
	Name            string              `yaml:"name"`
	Image           string              `yaml:"image"`
	TfxComponents   string              `yaml:"tfxComponents"`
	Env             map[string]string   `yaml:"env"`
	BeamArgs        map[string][]string `yaml:"beamArgs"`
}

func NamedValuesToMap(namedValues []apis.NamedValue) map[string]string {
	m := make(map[string]string)

	for _, nv := range namedValues {
		m[nv.Name] = nv.Value
	}

	return m
}

func NamedValuesToMultiMap(namedValues []apis.NamedValue) map[string][]string {
	multimap := make(map[string][]string)

	for _, nv := range namedValues {
		if _, found := multimap[nv.Name]; !found {
			multimap[nv.Name] = []string{}
		}

		multimap[nv.Name] = append(multimap[nv.Name], nv.Value)
	}

	return multimap
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

	beamArgs := append(wf.Config.DefaultBeamArgs, pipeline.Spec.BeamArgs...)
	beamArgs = append(beamArgs, apis.NamedValue{Name: "temp_location", Value: pipelineRoot + tempPath})

	return &CompilerConfig{
		RootLocation:    pipelineRoot,
		ServingLocation: pipelineRoot + servingPath,
		Name:            pipeline.ObjectMeta.Name,
		Image:           pipeline.Spec.Image,
		TfxComponents:   pipeline.Spec.TfxComponents,
		Env:             NamedValuesToMap(pipeline.Spec.Env),
		BeamArgs:        NamedValuesToMultiMap(beamArgs),
	}
}

func (workflows PipelineWorkflowFactory) ConstructCreationWorkflow(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	compilerConfigYaml, err := workflows.newCompilerConfig(pipeline).AsYaml()

	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(pipeline, WorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  PipelineWorkflowConstants.CompilerConfigParameterName,
						Value: argo.AnyStringPtr(compilerConfigYaml),
					},
					{
						Name:  PipelineWorkflowConstants.PipelineImageParameterName,
						Value: argo.AnyStringPtr(pipeline.Spec.Image),
					},
					{
						Name:  PipelineWorkflowConstants.PipelineNameParameterName,
						Value: argo.AnyStringPtr(pipeline.Name),
					},
					{
						Name:  PipelineWorkflowConstants.PipelineVersionParameterName,
						Value: argo.AnyStringPtr(pipeline.Spec.ComputeVersion()),
					},
					{
						Name:  WorkflowConstants.KfpEndpointParameterName,
						Value: argo.AnyStringPtr(workflows.Config.KfpEndpoint),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         workflows.Config.WorkflowTemplatePrefix + "create-pipeline",
				ClusterScope: true,
			},
		},
	}, nil
}

func (workflows PipelineWorkflowFactory) ConstructUpdateWorkflow(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	compilerConfigYaml, err := workflows.newCompilerConfig(pipeline).AsYaml()

	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(pipeline, WorkflowConstants.UpdateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  PipelineWorkflowConstants.CompilerConfigParameterName,
						Value: argo.AnyStringPtr(compilerConfigYaml),
					},
					{
						Name:  PipelineWorkflowConstants.PipelineIdParameterName,
						Value: argo.AnyStringPtr(pipeline.Status.KfpId),
					},
					{
						Name:  PipelineWorkflowConstants.PipelineImageParameterName,
						Value: argo.AnyStringPtr(pipeline.Spec.Image),
					},
					{
						Name:  PipelineWorkflowConstants.PipelineVersionParameterName,
						Value: argo.AnyStringPtr(pipeline.Spec.ComputeVersion()),
					},
					{
						Name:  WorkflowConstants.KfpEndpointParameterName,
						Value: argo.AnyStringPtr(workflows.Config.KfpEndpoint),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         workflows.Config.WorkflowTemplatePrefix + "update-pipeline",
				ClusterScope: true,
			},
		},
	}, nil
}

func (workflows PipelineWorkflowFactory) ConstructDeletionWorkflow(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(pipeline, WorkflowConstants.DeleteOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  PipelineWorkflowConstants.PipelineIdParameterName,
						Value: argo.AnyStringPtr(pipeline.Status.KfpId),
					},
					{
						Name:  WorkflowConstants.KfpEndpointParameterName,
						Value: argo.AnyStringPtr(workflows.Config.KfpEndpoint),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         workflows.Config.WorkflowTemplatePrefix + "delete-pipeline",
				ClusterScope: true,
			},
		},
	}, nil
}
