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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PipelineParamsCreator", func() {
	expectedEnv := []apis.NamedValue{
		{Name: "a", Value: "b"},
	}

	expectedBeamArgs := []apis.NamedValue{
		{Name: "c", Value: "d"},
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
			Framework:     "pipelineFramework",
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
				Expect(compilerConfig.Framework).To(Equal("pipelineFramework"))
				Expect(compilerConfig.Env).To(Equal(expectedEnv))
				Expect(compilerConfig.BeamArgs).To(Equal(expectedBeamArgs))
			})

			It("creates valid YAML", func() {
				configYaml, err := yaml.Marshal(compilerConfig)
				Expect(err).NotTo(HaveOccurred())

				m := make(map[interface{}]interface{})
				yaml.Unmarshal(configYaml, m)

				Expect(m["name"]).To(Equal("pipelineNamespace/pipelineName"))
				Expect(m["image"]).To(Equal("pipelineImage"))
				Expect(m["framework"]).To(Equal("pipelineFramework"))
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
						"pipelineFramework": expectedImage,
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
						"somethingElse": "something/else",
					},
				}
				creator := PipelineParamsCreator{Config: config}
				_, err := creator.additionalParams(pipeline)

				Expect(err.Error()).To(Equal("error in workflow: [pipelineFramework] framework not found"))
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
				pipeline.Spec.Framework = ""
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
				pipeline.Spec.Framework = ""
				_, err := creator.additionalParams(pipeline)

				Expect(err.Error()).To(Equal("error in workflow: [default] framework not found"))
			})
		})
	})
})
