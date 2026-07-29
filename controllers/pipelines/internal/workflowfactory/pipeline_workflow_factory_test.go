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

var _ = Describe("pipelineWorkflowFactory", func() {
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
			Payload: `[{"op": "replace", "path": "/image", "value": "patchedImage"}]`,
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
		compilerConfig := pipelineDefinition(pipeline)

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

			var result any
			err = json.Unmarshal(configYaml, &result)
			Expect(err).NotTo(HaveOccurred())
			resultMap, ok := result.(map[string]any)
			Expect(ok).To(BeTrue())

			Expect(resultMap["name"]).To(Equal("pipelineNamespace/pipelineName"))
			Expect(resultMap["image"]).To(Equal("pipelineImage"))

			Expect(resultMap["framework"]).NotTo(BeNil())

			framework, ok := resultMap["framework"].(map[string]any)
			Expect(ok).To(BeTrue())

			Expect(framework["name"]).To(Equal(expectedFramework.Name))

			parameters := framework["parameters"].(map[string]any)
			Expect(parameters["a"]).To(Equal("b"))
			Expect(parameters["c"]).To(Equal("d"))

			env := resultMap["env"].([]any)
			Expect(env[0]).To(Equal(map[string]any{
				"name":  "a",
				"value": "b",
			}))
		})
	})

	Context("creationParams", func() {
		factory := pipelineWorkflowFactory{}

		When("the Pipeline resource specifies a valid framework", func() {
			It("returns the definition and framework image parameters", func() {
				expectedImage := "registry/pipelineFramework"
				params, err := factory.creationParams(provider, pipeline)
				Expect(err).NotTo(HaveOccurred())

				Expect(params).To(HaveLen(2))
				Expect(params[0].Name).To(Equal(workflowconstants.ResourceDefinitionParameterName))
				Expect(params[1]).To(Equal(argo.Parameter{
					Name:  workflowconstants.PipelineFrameworkImageParameterName,
					Value: argo.AnyStringPtr(expectedImage),
				}))
			})

			It("applies the framework patches to the definition", func() {
				params, err := factory.creationParams(provider, pipeline)
				Expect(err).NotTo(HaveOccurred())

				var definition map[string]interface{}
				Expect(json.Unmarshal([]byte(params[0].Value.String()), &definition)).To(Succeed())
				Expect(definition["image"]).To(Equal("patchedImage"))
			})
		})

		When("the requested framework is not found", func() {
			It("returns an error", func() {
				_, err := factory.creationParams(provider, pipelineIncorrectFramework)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error in workflow: [invalidFramework] framework not support by provider"))
			})
		})
	})
})
