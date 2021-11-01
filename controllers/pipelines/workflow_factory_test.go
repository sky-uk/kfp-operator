package pipelines

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("WorkflowFactory.Annotations", func() {
	When("no debug annotation is present on the ObjectMeta", func() {
		It("uses the configured defaults", func() {
			workflowFactory := WorkflowFactory{
				Config: configv1.Configuration{
					Debug: pipelinesv1.DebugOptions{
						KeepWorkflows: true,
					},
				},
			}

			objectMeta := metav1.ObjectMeta{Annotations: map[string]string{}}

			debugAnnotations := workflowFactory.Annotations(context.Background(), objectMeta)["pipelines.kubeflow.org/debug"]
			Expect(debugAnnotations).To(MatchJSON(`{"keepWorkflows":true}`))
		})
	})

	When("the annotation can't be deserialized", func() {
		It("uses the configured defaults", func() {
			workflowFactory := WorkflowFactory{
				Config: configv1.Configuration{
					Debug: pipelinesv1.DebugOptions{
						KeepWorkflows: true,
					},
				},
			}

			objectMeta := metav1.ObjectMeta{Annotations: map[string]string{
				"pipelines.kubeflow.org/debug": "broken:",
			}}

			debugAnnotations := workflowFactory.Annotations(context.Background(), objectMeta)["pipelines.kubeflow.org/debug"]
			Expect(debugAnnotations).To(MatchJSON(`{"keepWorkflows":true}`))
		})
	})

	When("the annotation is valid and ObjectMeta overrides the default config", func() {
		It("uses the overrides", func() {
			workflowFactory := WorkflowFactory{
				Config: configv1.Configuration{
					Debug: pipelinesv1.DebugOptions{
						KeepWorkflows: false,
					},
				},
			}

			objectMeta := metav1.ObjectMeta{Annotations: map[string]string{
				"pipelines.kubeflow.org/debug": `{"keepWorkflows":true}`,
			}}

			debugAnnotations := workflowFactory.Annotations(context.Background(), objectMeta)["pipelines.kubeflow.org/debug"]
			Expect(debugAnnotations).To(MatchJSON(`{"keepWorkflows":true}`))
		})
	})

	When("the annotation is valid and the ObjectMeta does not override the default config", func() {
		It("uses the configured defaults", func() {
			workflowFactory := WorkflowFactory{
				Config: configv1.Configuration{
					Debug: pipelinesv1.DebugOptions{
						KeepWorkflows: true,
					},
				},
			}

			objectMeta := metav1.ObjectMeta{Annotations: map[string]string{
				"pipelines.kubeflow.org/debug": `{"keepWorkflows":false}`,
			}}

			debugAnnotations := workflowFactory.Annotations(context.Background(), objectMeta)["pipelines.kubeflow.org/debug"]
			Expect(debugAnnotations).To(MatchJSON(`{"keepWorkflows":true}`))
		})
	})
})
