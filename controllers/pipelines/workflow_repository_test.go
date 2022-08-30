package pipelines

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha3"
)

var _ = Describe("WorkflowRepository.debugAnnotations", func() {
	When("no debug annotation is present on the ObjectMeta", func() {
		It("uses the configured defaults", func() {
			workflowRepository := WorkflowRepositoryImpl{
				Config: config.Configuration{
					Debug: apis.DebugOptions{
						KeepWorkflows: true,
					},
				},
			}

			annotations := map[string]string{}

			debugAnnotations := workflowRepository.debugAnnotations(context.Background(), annotations)["pipelines.kubeflow.org/debug"]
			Expect(debugAnnotations).To(MatchJSON(`{"keepWorkflows":true}`))
		})
	})

	When("the annotation can't be deserialized", func() {
		It("uses the configured defaults", func() {
			workflowRepository := WorkflowRepositoryImpl{
				Config: config.Configuration{
					Debug: apis.DebugOptions{
						KeepWorkflows: true,
					},
				},
			}

			annotations := map[string]string{
				"pipelines.kubeflow.org/debug": "broken:",
			}

			debugAnnotations := workflowRepository.debugAnnotations(context.Background(), annotations)["pipelines.kubeflow.org/debug"]
			Expect(debugAnnotations).To(MatchJSON(`{"keepWorkflows":true}`))
		})
	})

	DescribeTable("the annotation is valid", func(enabledInAnnotation bool, enabledInConfig bool, expectation string) {
		workflowRepository := WorkflowRepositoryImpl{
			Config: config.Configuration{
				Debug: apis.DebugOptions{
					KeepWorkflows: enabledInConfig,
				},
			},
		}

		annotations := map[string]string{
			"pipelines.kubeflow.org/debug": fmt.Sprintf(`{"keepWorkflows":%t}`, enabledInAnnotation),
		}

		debugAnnotations := workflowRepository.debugAnnotations(context.Background(), annotations)["pipelines.kubeflow.org/debug"]
		Expect(debugAnnotations).To(MatchJSON(expectation))
	},
		Entry("enabled in annotation and config", true, true, `{"keepWorkflows":true}`),
		Entry("enabled in annotation but disabled in config", true, false, `{"keepWorkflows":true}`),
		Entry("disabled in annotation but enabled in config", false, true, `{"keepWorkflows":true}`),
		Entry("disabled in annotation and config", false, false, `{}`),
	)
})
