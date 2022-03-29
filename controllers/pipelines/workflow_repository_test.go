package pipelines

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

var _ = Describe("WorkflowRepository.Annotations", func() {
	When("no debug annotation is present on the ObjectMeta", func() {
		It("uses the configured defaults", func() {
			workflowRepository := WorkflowRepositoryImpl{
				Config: configv1.Configuration{
					Debug: pipelinesv1.DebugOptions{
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
				Config: configv1.Configuration{
					Debug: pipelinesv1.DebugOptions{
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

	When("the annotation is valid and ObjectMeta overrides the default config", func() {
		It("uses the overrides", func() {
			workflowRepository := WorkflowRepositoryImpl{
				Config: configv1.Configuration{
					Debug: pipelinesv1.DebugOptions{
						KeepWorkflows: false,
					},
				},
			}

			annotations := map[string]string{
				"pipelines.kubeflow.org/debug": `{"keepWorkflows":true}`,
			}

			debugAnnotations := workflowRepository.debugAnnotations(context.Background(), annotations)["pipelines.kubeflow.org/debug"]
			Expect(debugAnnotations).To(MatchJSON(`{"keepWorkflows":true}`))
		})
	})

	When("the annotation is valid and the ObjectMeta does not override the default config", func() {
		It("uses the configured defaults", func() {
			workflowRepository := WorkflowRepositoryImpl{
				Config: configv1.Configuration{
					Debug: pipelinesv1.DebugOptions{
						KeepWorkflows: true,
					},
				},
			}

			annotations := map[string]string{
				"pipelines.kubeflow.org/debug": `{"keepWorkflows":false}`,
			}

			debugAnnotations := workflowRepository.debugAnnotations(context.Background(), annotations)["pipelines.kubeflow.org/debug"]
			Expect(debugAnnotations).To(MatchJSON(`{"keepWorkflows":true}`))
		})
	})
})
