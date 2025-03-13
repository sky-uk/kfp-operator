//go:build unit

package workflowfactory

import (
	"encoding/json"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PipelineParamsCreator", func() {
	expectedEnv := []apis.NamedValue{
		{Name: "a", Value: "b"},
	}

	expectedFramework := pipelineshub.PipelineFramework{
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
			Image:     "pipelineImage",
			Env:       expectedEnv,
			Framework: expectedFramework,
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
				Expect(compilerConfig.Framework).To(Equal(expectedFramework))
				Expect(compilerConfig.Env).To(Equal(expectedEnv))
			})

			It("creates valid JSON", func() {
				configYaml, err := json.Marshal(compilerConfig)
				Expect(err).NotTo(HaveOccurred())

				var result interface{}
				err = json.Unmarshal(configYaml, &result)
				Expect(err).NotTo(HaveOccurred())
				resultMap, ok := result.(map[string]interface{})
				Expect(ok).To(BeTrue())

				Expect(resultMap["name"]).To(Equal("pipelineNamespace/pipelineName"))
				Expect(resultMap["image"]).To(Equal("pipelineImage"))

				framework := resultMap["framework"].(map[string]interface{})
				Expect(framework["type"]).To(Equal(expectedFramework.Type))

				parameters := framework["parameters"].(map[string]interface{})
				Expect(parameters["a"]).To(Equal("b"))
				Expect(parameters["c"]).To(Equal("d"))

				env := resultMap["env"].([]interface{})
				Expect(env[0]).To(Equal(map[string]interface{}{
					"name":  "a",
					"value": "b",
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

	})
})
