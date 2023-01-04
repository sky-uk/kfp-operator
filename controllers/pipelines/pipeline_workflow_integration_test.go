//go:build integration
// +build integration

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Context("Resource Workflows", Serial, func() {
	workflowFactory := PipelineWorkflowFactory(config.Configuration{
		DefaultExperiment:      "Default",
		DefaultProvider:        "stub",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newPipeline = func() *pipelinesv1.Pipeline {
		return &pipelinesv1.Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name:      apis.RandomLowercaseString(),
				Namespace: "argo",
			},
			Spec: pipelinesv1.PipelineSpec{
				Image: "kfp-operator-argo-kfp-pipeline",
			},
			Status: pipelinesv1.Status{
				ProviderId: pipelinesv1.ProviderAndId{
					Provider: "stub",
					Id:       apis.RandomString(),
				},
			},
		}
	}

	DescribeTable("Pipeline Workflows", AssertWorkflow[*pipelinesv1.Pipeline],
		Entry("Creation",
			newPipeline,
			StubWithIdAndError[*pipelinesv1.Pipeline],
			workflowFactory.ConstructCreationWorkflow,
		), Entry("Update",
			newPipeline,
			StubWithIdAndError[*pipelinesv1.Pipeline],
			workflowFactory.ConstructUpdateWorkflow,
		), Entry("Deletion succeeds",
			newPipeline,
			StubWithEmpty[*pipelinesv1.Pipeline],
			workflowFactory.ConstructDeletionWorkflow,
		), Entry("Deletion fails",
			newPipeline,
			StubWithExistingIdAndError[*pipelinesv1.Pipeline],
			workflowFactory.ConstructDeletionWorkflow,
		),
	)
})
