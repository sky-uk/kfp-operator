//go:build unit
// +build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1alpha2"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha2"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PipelineConfig", func() {

	Specify("Some fields are copied from Pipeline resource", func() {
		wf := PipelineWorkflowFactory{}
		pipeline := &pipelinesv1.Pipeline{
			ObjectMeta: v1.ObjectMeta{
				Name: "pipelineName",
			},
			Spec: pipelinesv1.PipelineSpec{
				Image:         "pipelineImage",
				TfxComponents: "pipelineTfxComponents",
				Env: map[string]string{
					"ea": "b",
				},
			},
		}

		compilerConfig := wf.newCompilerConfig(pipeline)

		Expect(compilerConfig.Name).To(Equal("pipelineName"))
		Expect(compilerConfig.Image).To(Equal("pipelineImage"))
		Expect(compilerConfig.TfxComponents).To(Equal("pipelineTfxComponents"))
		Expect(compilerConfig.Env["ea"]).To(Equal("b"))
	})

	Specify("Paths are appended to PipelineStorage", func() {
		wf := PipelineWorkflowFactory{
			WorkflowFactoryBase: WorkflowFactoryBase{
				Config: configv1.Configuration{
					PipelineStorage: "gs://bucket",
				},
			},
		}
		pipeline := &pipelinesv1.Pipeline{
			ObjectMeta: v1.ObjectMeta{
				Name: "pipelineName",
			},
		}

		compilerConfig := wf.newCompilerConfig(pipeline)

		Expect(compilerConfig.RootLocation).To(Equal("gs://bucket/pipelineName"))
		Expect(compilerConfig.ServingLocation).To(Equal("gs://bucket/pipelineName/serving"))
	})

	Specify("Original BeamArgs are copied", func() {
		wf := PipelineWorkflowFactory{}
		pipeline := &pipelinesv1.Pipeline{
			Spec: pipelinesv1.PipelineSpec{
				BeamArgs: []pipelinesv1.NamedValue{
					{Name: "a", Value: "b"},
				},
			},
		}

		compilerConfig := wf.newCompilerConfig(pipeline)

		Expect(compilerConfig.BeamArgs["a"]).To(Equal([]string{"b"}))
	})

	Specify("BeamArgs are enriched with temp_location", func() {
		wf := PipelineWorkflowFactory{
			WorkflowFactoryBase: WorkflowFactoryBase{
				Config: configv1.Configuration{
					PipelineStorage: "gs://bucket",
				},
			},
		}

		pipeline := &pipelinesv1.Pipeline{
			ObjectMeta: v1.ObjectMeta{
				Name: "pipelineName",
			},
		}

		compilerConfig := wf.newCompilerConfig(pipeline)

		Expect(compilerConfig.BeamArgs["temp_location"]).To(Equal([]string{"gs://bucket/pipelineName/tmp"}))
	})

	Specify("BeamArgs are appended to default configuration values", func() {
		wf := PipelineWorkflowFactory{
			WorkflowFactoryBase: WorkflowFactoryBase{
				Config: configv1.Configuration{
					DefaultBeamArgs: []pipelinesv1.NamedValue{
						{Name: "ba", Value: "default"},
						{Name: "bc", Value: "default"},
					},
				},
			},
		}

		pipeline := &pipelinesv1.Pipeline{
			Spec: pipelinesv1.PipelineSpec{
				BeamArgs: []pipelinesv1.NamedValue{
					{Name: "bc", Value: "bd"},
				},
			},
		}

		compilerConfig := wf.newCompilerConfig(pipeline)

		Expect(compilerConfig.BeamArgs["ba"]).To(Equal([]string{"default"}))
		Expect(compilerConfig.BeamArgs["bc"]).To(Equal([]string{"default", "bd"}))
	})

	It("Creates a valid YAML", func() {
		config := CompilerConfig{
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

		configYaml, err := config.AsYaml()
		Expect(err).NotTo(HaveOccurred())

		m := make(map[interface{}]interface{})
		yaml.Unmarshal([]byte(configYaml), m)

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
