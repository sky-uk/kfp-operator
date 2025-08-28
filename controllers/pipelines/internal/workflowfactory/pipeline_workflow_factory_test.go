//go:build unit

package workflowfactory

import (
	"encoding/json"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	"github.com/sky-uk/kfp-operator/pkg/common"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

var _ = Describe("PipelineParamsCreator", func() {
	expectedEnv := []apis.NamedValue{
		{Name: "a", Value: "b"},
	}

	expectedFramework := pipelineshub.PipelineFramework{
		Name: "pipelineFramework",
		Parameters: map[string]*apiextensionsv1.JSON{
			"a": {Raw: []byte(`"b"`)},
			"c": {Raw: []byte(`"d"`)},
		},
	}

	expectedPatches := []pipelineshub.Patch{
		{
			Type:    "json",
			Payload: `{"op": "add", "path": "/spec/parameters", "value": {"a": "b"}}`,
		},
	}

	provider := pipelineshub.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "providerName",
			Namespace: "providerNamespace",
		},
		Spec: pipelineshub.ProviderSpec{
			Frameworks: []pipelineshub.Framework{
				{
					Name:    strings.ToLower(expectedFramework.Name),
					Image:   "registry/pipelineFramework",
					Patches: expectedPatches,
				},
			},
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

	pipelineIncorrectFramework := pipeline.DeepCopy()
	pipelineIncorrectFramework.Spec.Framework.Name = "invalidFramework"

	Context("pipelineDefinition", func() {
		creator := PipelineParamsCreator{}

		When("given a Pipeline resource with invalid framework", func() {
			_, _, err := creator.pipelineDefinition(provider, pipelineIncorrectFramework)
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error in workflow: [invalidFramework] framework not support by provider"))
			})
		})

		When("given a Pipeline resource with valid framework", func() {
			patches, compilerConfig, _ := creator.pipelineDefinition(provider, pipeline)
			It("creates a valid PipelineDefinition", func() {
				Expect(compilerConfig.Name).To(Equal(common.NamespacedName{
					Name:      "pipelineName",
					Namespace: "pipelineNamespace",
				}))
				Expect(compilerConfig.Image).To(Equal("pipelineImage"))
				Expect(compilerConfig.Framework).To(Equal(expectedFramework))
				Expect(compilerConfig.Env).To(Equal(expectedEnv))
				Expect(patches).To(Equal(expectedPatches))
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

				Expect(resultMap["framework"]).NotTo(BeNil())

				framework, ok := resultMap["framework"].(map[string]interface{})
				Expect(ok).To(BeTrue())

				Expect(framework["name"]).To(Equal(expectedFramework.Name))

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
				creator := PipelineParamsCreator{}
				params, err := creator.additionalParams(provider, pipeline)
				Expect(err).NotTo(HaveOccurred())
				Expect(params).To(Equal([]argo.Parameter{
					{
						Name:  workflowconstants.PipelineFrameworkImageParameterName,
						Value: argo.AnyStringPtr(expectedImage),
					},
				}))
			})

			It("returns an error if the requested framework is not found", func() {
				creator := PipelineParamsCreator{}
				_, err := creator.additionalParams(provider, pipelineIncorrectFramework)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error in workflow: [invalidFramework] framework not support by provider"))
			})
		})

	})
})
