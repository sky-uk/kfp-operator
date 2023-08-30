//go:build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
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

	Specify("Paths are appended to PipelineStorage", func() {
		wf := PipelineDefinitionCreator{
			Config: config.Configuration{
				PipelineStorage: "gs://bucket",
			},
		}
		pipeline := &pipelinesv1.Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pipelineName",
			},
		}

		compilerConfig, _ := wf.pipelineDefinition(pipeline)

		Expect(compilerConfig.RootLocation).To(Equal("gs://bucket/pipelineName"))
		Expect(compilerConfig.ServingLocation).To(Equal("gs://bucket/pipelineName/serving"))
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

	Specify("BeamArgs are enriched with temp_location", func() {
		wf := PipelineDefinitionCreator{
			Config: config.Configuration{
				PipelineStorage: "gs://bucket",
			},
		}

		pipeline := &pipelinesv1.Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pipelineName",
			},
		}

		compilerConfig, _ := wf.pipelineDefinition(pipeline)

		Expect(compilerConfig.BeamArgs["temp_location"]).To(Equal([]string{"gs://bucket/pipelineName/tmp"}))
	})

	Specify("BeamArgs are appended to default configuration values", func() {
		wf := PipelineDefinitionCreator{
			Config: config.Configuration{
				DefaultBeamArgs: []apis.NamedValue{
					{Name: "ba", Value: "default"},
					{Name: "bc", Value: "default"},
				},
			},
		}

		pipeline := &pipelinesv1.Pipeline{
			Spec: pipelinesv1.PipelineSpec{
				BeamArgs: []apis.NamedValue{
					{Name: "bc", Value: "bd"},
				},
			},
		}

		compilerConfig, _ := wf.pipelineDefinition(pipeline)

		Expect(compilerConfig.BeamArgs["ba"]).To(Equal([]string{"default"}))
		Expect(compilerConfig.BeamArgs["bc"]).To(Equal([]string{"default", "bd"}))
	})

	It("Creates a valid YAML", func() {
		config := providers.PipelineDefinition{
			RootLocation:    "pipelineRootLocation",
			ServingLocation: "pipelineServingLocation",
			Name:            "pipelineName",
			Image:           "pipelineImage",
			TfxComponents:   "pipelineTfxComponents",
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

		Expect(m["rootLocation"]).To(Equal("pipelineRootLocation"))
		Expect(m["servingLocation"]).To(Equal("pipelineServingLocation"))
		Expect(m["name"]).To(Equal("pipelineName"))
		Expect(m["image"]).To(Equal("pipelineImage"))
		Expect(m["tfxComponents"]).To(Equal("pipelineTfxComponents"))
		env := m["env"].(map[interface{}]interface{})
		Expect(env["ea"]).To(Equal("eb"))
		beamArgs := m["beamArgs"].(map[interface{}]interface{})
		Expect(beamArgs["ba"]).To(Equal([]interface{}{"bb"}))
	})
})
