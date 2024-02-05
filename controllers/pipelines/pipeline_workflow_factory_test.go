//go:build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PipelineDefinition", func() {

	Specify("Some fields are copied from Pipeline resource", func() {
		wf := PipelineDefinitionCreator{}
		pipeline := &pipelinesv1.Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pipelineName",
			},
			Spec: pipelinesv1.PipelineSpec{
				Image:         "pipelineImage",
				TfxComponents: "pipelineTfxComponents",
				Env: []apis.NamedValue{
					{Name: "ea", Value: "b"},
				},
			},
		}

		compilerConfig, _ := wf.pipelineDefinition(pipeline)

		Expect(compilerConfig.Name).To(Equal("pipelineName"))
		Expect(compilerConfig.Image).To(Equal("pipelineImage"))
		Expect(compilerConfig.TfxComponents).To(Equal("pipelineTfxComponents"))
		Expect(compilerConfig.Env["ea"]).To(Equal("b"))
	})

	Specify("Original BeamArgs are copied", func() {
		wf := PipelineDefinitionCreator{}
		pipeline := &pipelinesv1.Pipeline{
			Spec: pipelinesv1.PipelineSpec{
				BeamArgs: []apis.NamedValue{
					{Name: "a", Value: "b"},
				},
			},
		}

		compilerConfig, _ := wf.pipelineDefinition(pipeline)

		Expect(compilerConfig.BeamArgs["a"]).To(Equal([]string{"b"}))
	})

	It("Creates a valid YAML", func() {
		config := providers.PipelineDefinition{
			Name:          "pipelineName",
			Image:         "pipelineImage",
			TfxComponents: "pipelineTfxComponents",
			Env: map[string]string{
				"ea": "eb",
			},
			BeamArgs: map[string][]string{
				"ba": {"bb"},
			},
		}

		configYaml, err := yaml.Marshal(config)
		Expect(err).NotTo(HaveOccurred())

		m := make(map[interface{}]interface{})
		yaml.Unmarshal(configYaml, m)

		Expect(m["name"]).To(Equal("pipelineName"))
		Expect(m["image"]).To(Equal("pipelineImage"))
		Expect(m["tfxComponents"]).To(Equal("pipelineTfxComponents"))
		env := m["env"].(map[interface{}]interface{})
		Expect(env["ea"]).To(Equal("eb"))
		beamArgs := m["beamArgs"].(map[interface{}]interface{})
		Expect(beamArgs["ba"]).To(Equal([]interface{}{"bb"}))
	})
})
