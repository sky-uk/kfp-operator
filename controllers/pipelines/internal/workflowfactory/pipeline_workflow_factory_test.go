//go:build unit

package workflowfactory

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	"gopkg.in/yaml.v2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func extractParameterRaw(framework map[interface{}]interface{}, parameterKey string) []interface{} {
	parameters := framework["parameters"].(map[interface{}]interface{})

	return parameters[parameterKey].(map[interface{}]interface{})["raw"].([]interface{})
}

func convertInterfaceArrayToString(values []interface{}) string {
	var result []byte
	for _, v := range values {
		result = append(result, byte(v.(int))) // Convert int to byte
	}
	return string(result)
}

var _ = Describe("PipelineParamsCreator", func() {
	expectedEnv := []apis.NamedValue{
		{Name: "a", Value: "b"},
	}

	expectedBeamArgs := []apis.NamedValue{
		{Name: "c", Value: "d"},
	}

	expectedFramework := &pipelineshub.PipelineFramework{
		Type: "pipelineFramework",
		Parameters: map[string]*apiextensionsv1.JSON{
			"a": {Raw: []byte(`"b"`)},
			"c": {Raw: []byte(`"d"`)},
		},
	}

	pipeline := &pipelineshub.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pipelineName",
			Namespace: "pipelineNamespace",
		},
		Spec: pipelineshub.PipelineSpec{
			Image:         "pipelineImage",
			TfxComponents: "pipelineTfxComponents",
			Env:           expectedEnv,
			BeamArgs:      expectedBeamArgs,
			Framework:     expectedFramework,
		},
	}
	Context("pipelineDefinition", func() {
		creator := PipelineParamsCreator{}

		compilerConfig, _ := creator.pipelineDefinition(pipeline)
		When("given a Pipeline resource", func() {
			It("creates a valid PipelineDefinition", func() {
				Expect(compilerConfig.Name).To(Equal(common.NamespacedName{
					Name:      "pipelineName",
					Namespace: "pipelineNamespace",
				}))
				Expect(compilerConfig.Image).To(Equal("pipelineImage"))
				Expect(compilerConfig.TfxComponents).To(Equal("pipelineTfxComponents"))
				Expect(compilerConfig.Framework).To(Equal(expectedFramework))
				Expect(compilerConfig.Env).To(Equal(expectedEnv))
				Expect(compilerConfig.BeamArgs).To(Equal(expectedBeamArgs))
			})

			It("creates valid YAML", func() {
				configYaml, err := yaml.Marshal(compilerConfig)
				Expect(err).NotTo(HaveOccurred())

				m := make(map[interface{}]interface{})
				err = yaml.Unmarshal(configYaml, m)
				Expect(err).NotTo(HaveOccurred())

				Expect(m["name"]).To(Equal("pipelineNamespace/pipelineName"))
				Expect(m["image"]).To(Equal("pipelineImage"))

				framework := m["framework"].(map[interface{}]interface{})
				Expect(framework["type"]).To(Equal(expectedFramework.Type))

				aValue := extractParameterRaw(framework, "a")
				bValue := extractParameterRaw(framework, "c")

				Expect(convertInterfaceArrayToString(aValue)).To(Equal("\"b\""))
				Expect(convertInterfaceArrayToString(bValue)).To(Equal("\"d\""))

				Expect(m["tfxComponents"]).To(Equal("pipelineTfxComponents"))
				env := m["env"].([]interface{})
				Expect(env[0]).To(Equal(map[interface{}]interface{}{
					"name":  "a",
					"value": "b",
				}))
				beamArgs := m["beamArgs"].([]interface{})
				Expect(beamArgs[0]).To(Equal(map[interface{}]interface{}{
					"name":  "c",
					"value": "d",
				}))

			})
		})
	})

	Context("additionalParams", func() {
		When("the Pipeline resource specifies a framework", func() {
			It("returns additional pipeline framework image parameter for the framework requested", func() {
				expectedImage := "registry/pipelineFramework"
				config := config.KfpControllerConfigSpec{
					PipelineFrameworkImages: map[string]string{
						"pipelineframework": expectedImage,
					},
				}
				creator := PipelineParamsCreator{Config: config}
				params, err := creator.additionalParams(pipeline)
				Expect(err).NotTo(HaveOccurred())
				Expect(params).To(Equal([]argo.Parameter{
					{
						Name:  workflowconstants.PipelineFrameworkImageParameterName,
						Value: argo.AnyStringPtr(expectedImage),
					},
				}))
			})

			It("returns an error if the requested framework is not found", func() {
				config := config.KfpControllerConfigSpec{
					PipelineFrameworkImages: map[string]string{
						"somethingelse": "something/else",
					},
				}
				creator := PipelineParamsCreator{Config: config}
				_, err := creator.additionalParams(pipeline)

				Expect(err.Error()).To(Equal("error in workflow: [pipelineframework] framework not found"))
			})
		})

		When("the Pipeline resource does not specify a framework", func() {
			It("returns additional pipeline framework image parameter for the default framework", func() {
				expectedImage := "registry/default"
				config := config.KfpControllerConfigSpec{
					PipelineFrameworkImages: map[string]string{
						"default": expectedImage,
					},
				}
				creator := PipelineParamsCreator{Config: config}
				pipeline.Spec.Framework = nil
				params, err := creator.additionalParams(pipeline)
				Expect(err).NotTo(HaveOccurred())
				Expect(params).To(Equal([]argo.Parameter{
					{
						Name:  workflowconstants.PipelineFrameworkImageParameterName,
						Value: argo.AnyStringPtr(expectedImage),
					},
				}))
			})

			It("returns an error if no default framework is set", func() {
				config := config.KfpControllerConfigSpec{
					PipelineFrameworkImages: map[string]string{},
				}
				creator := PipelineParamsCreator{Config: config}
				pipeline.Spec.Framework = nil
				_, err := creator.additionalParams(pipeline)

				Expect(err.Error()).To(Equal("error in workflow: [default] framework not found"))
			})
		})
	})
})
