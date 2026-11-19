//go:build unit

package provider

import (
	"github.com/kubeflow/pipelines/api/v2alpha1/go/pipelinespec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/structpb"
)

var _ = Describe("LabelService", func() {
	Context("InsertLabelsIntoParameters", func() {
		expectLabelParameter := func(parameter *pipelinespec.ComponentInputsSpec_ParameterSpec) {
			GinkgoHelper()
			Expect(parameter.GetParameterType()).To(Equal(pipelinespec.ParameterType_STRING))
			Expect(parameter.GetIsOptional()).To(BeTrue())
			Expect(parameter.GetDefaultValue().AsInterface()).To(Equal(""))
		}

		It("inserts each label as an optional string parameter of the root component", func() {
			pipelineSpec := &pipelinespec.PipelineSpec{}

			NewDefaultLabelService().InsertLabelsIntoParameters(pipelineSpec, []string{"label1", "label2"})

			parameters := pipelineSpec.GetRoot().GetInputDefinitions().GetParameters()
			Expect(parameters).To(HaveLen(2))
			expectLabelParameter(parameters["label1"])
			expectLabelParameter(parameters["label2"])
		})

		It("keeps other parameters", func() {
			pipelineSpec := &pipelinespec.PipelineSpec{Root: &pipelinespec.ComponentSpec{
				InputDefinitions: &pipelinespec.ComponentInputsSpec{
					Parameters: map[string]*pipelinespec.ComponentInputsSpec_ParameterSpec{
						"epochs": {ParameterType: pipelinespec.ParameterType_NUMBER_INTEGER},
					},
				},
			}}

			NewDefaultLabelService().InsertLabelsIntoParameters(pipelineSpec, []string{"label1"})

			parameters := pipelineSpec.GetRoot().GetInputDefinitions().GetParameters()
			Expect(parameters).To(HaveKey("epochs"))
			expectLabelParameter(parameters["label1"])
		})

		It("replaces an existing parameter with the same name", func() {
			pipelineSpec := &pipelinespec.PipelineSpec{Root: &pipelinespec.ComponentSpec{
				InputDefinitions: &pipelinespec.ComponentInputsSpec{
					Parameters: map[string]*pipelinespec.ComponentInputsSpec_ParameterSpec{
						"label1": {
							ParameterType: pipelinespec.ParameterType_NUMBER_INTEGER,
							DefaultValue:  structpb.NewNumberValue(1),
						},
					},
				},
			}}

			NewDefaultLabelService().InsertLabelsIntoParameters(pipelineSpec, []string{"label1"})

			expectLabelParameter(pipelineSpec.GetRoot().GetInputDefinitions().GetParameters()["label1"])
		})
	})
})
