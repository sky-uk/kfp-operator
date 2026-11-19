package provider

import (
	"github.com/kubeflow/pipelines/api/v2alpha1/go/pipelinespec"
	"google.golang.org/protobuf/types/known/structpb"
)

// LabelService inserts tracking labels into a pipeline spec's parameters.
type LabelService interface {
	InsertLabelsIntoParameters(pipelineSpec *pipelinespec.PipelineSpec, labels []string)
}

// DefaultLabelService inserts labels as optional string parameters of the
// pipeline's root component.
type DefaultLabelService struct{}

func NewDefaultLabelService() LabelService {
	return DefaultLabelService{}
}

// InsertLabelsIntoParameters adds each label to root.inputDefinitions.parameters
// as an optional STRING parameter with an empty default, replacing any existing
// parameter with the same name.
func (DefaultLabelService) InsertLabelsIntoParameters(pipelineSpec *pipelinespec.PipelineSpec, labels []string) {
	if pipelineSpec.Root == nil {
		pipelineSpec.Root = &pipelinespec.ComponentSpec{}
	}
	if pipelineSpec.Root.InputDefinitions == nil {
		pipelineSpec.Root.InputDefinitions = &pipelinespec.ComponentInputsSpec{}
	}
	if pipelineSpec.Root.InputDefinitions.Parameters == nil {
		pipelineSpec.Root.InputDefinitions.Parameters = map[string]*pipelinespec.ComponentInputsSpec_ParameterSpec{}
	}

	for _, label := range labels {
		pipelineSpec.Root.InputDefinitions.Parameters[label] = &pipelinespec.ComponentInputsSpec_ParameterSpec{
			ParameterType: pipelinespec.ParameterType_STRING,
			IsOptional:    true,
			DefaultValue:  structpb.NewStringValue(""),
		}
	}
}
