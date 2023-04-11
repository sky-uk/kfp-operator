//go:build integration
// +build integration

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
)

var _ = Context("Resource Workflows", Serial, func() {
	workflowFactory := RunConfigurationWorkflowFactory(config.Configuration{
		DefaultExperiment:      "Default",
		DefaultProvider:        "stub",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newRunConfiguration = func() *pipelinesv1.RunConfiguration {
		return withIntegrationTestFields(pipelinesv1.RandomRunConfiguration())
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
