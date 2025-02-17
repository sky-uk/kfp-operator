//go:build unit

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PipelineDefinition", func() {

	Specify("Some fields are copied from Pipeline resource", func() {
		wf := PipelineDefinitionCreator{}

		expectedEnv := []apis.NamedValue{
			{Name: "ea", Value: "b"},
		}
		expectedBeamArgs := []apis.NamedValue{
			{Name: "a", Value: "b"},
		}

		pipeline := &pipelinesv1.Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pipelineName",
				Namespace: "pipelineNamespace",
			},
			Spec: pipelinesv1.PipelineSpec{
				Image:         "pipelineImage",
				TfxComponents: "pipelineTfxComponents",
				Env:           expectedEnv,
				BeamArgs:      expectedBeamArgs,
				Framework:     "pipelineFramework",
			},
		}

		compilerConfig, _ := wf.pipelineDefinition(pipeline)

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

	It("Creates a valid YAML", func() {
		config := providers.PipelineDefinition{
			Name: common.NamespacedName{
				Name:      "pipelineName",
				Namespace: "pipelineNamespace",
			},
			Image:         "pipelineImage",
			TfxComponents: "pipelineTfxComponents",
			Framework:     "pipelineFramework",
			Env: []apis.NamedValue{
				{Name: "ea", Value: "eb"},
			},
			BeamArgs: []apis.NamedValue{
				{Name: "ba", Value: "bb"},
			},
		}

		configYaml, err := yaml.Marshal(config)
		Expect(err).NotTo(HaveOccurred())

		m := make(map[interface{}]interface{})
		yaml.Unmarshal(configYaml, m)

		Expect(m["name"]).To(Equal("pipelineNamespace/pipelineName"))
		Expect(m["image"]).To(Equal("pipelineImage"))
		Expect(m["framework"]).To(Equal("pipelineFramework"))
		Expect(m["tfxComponents"]).To(Equal("pipelineTfxComponents"))
		env := m["env"].([]interface{})
		Expect(env[0]).To(Equal(map[interface{}]interface{}{
			"name":  "ea",
			"value": "eb",
		}))
		beamArgs := m["beamArgs"].([]interface{})
		Expect(beamArgs[0]).To(Equal(map[interface{}]interface{}{
			"name":  "ba",
			"value": "bb",
		}))
	})
})
