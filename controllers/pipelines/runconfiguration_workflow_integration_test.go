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
	workflowFactory := RunConfigurationWorkflowFactory(config.Configuration{
		DefaultExperiment:      "Default",
		DefaultProvider:        "stub",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newRunConfiguration = func() *pipelinesv1.RunConfiguration {
		return &pipelinesv1.RunConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name:      apis.RandomLowercaseString(),
				Namespace: "argo",
			},
			Status: pipelinesv1.RunConfigurationStatus{
				ObservedPipelineVersion: apis.RandomString(),
				Status: pipelinesv1.Status{
					ProviderId: pipelinesv1.ProviderAndId{
						Provider: "stub",
						Id:       apis.RandomString(),
					},
				},
			},
		}
	}

	DescribeTable("RunConfiguration Workflows", AssertWorkflow[*pipelinesv1.RunConfiguration],
		Entry("Creation",
			newRunConfiguration,
			StubWithIdAndError[*pipelinesv1.RunConfiguration],
			workflowFactory.ConstructCreationWorkflow,
		), Entry("Update",
			newRunConfiguration,
			StubWithIdAndError[*pipelinesv1.RunConfiguration],
			workflowFactory.ConstructUpdateWorkflow,
		), Entry("Deletion succeeds",
			newRunConfiguration,
			StubWithEmpty[*pipelinesv1.RunConfiguration],
			workflowFactory.ConstructDeletionWorkflow,
		), Entry("Deletion fails",
			newRunConfiguration,
			StubWithExistingIdAndError[*pipelinesv1.RunConfiguration],
			workflowFactory.ConstructDeletionWorkflow,
		),
	)
})
