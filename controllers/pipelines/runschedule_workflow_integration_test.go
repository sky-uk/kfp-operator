//go:build integration
// +build integration

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
)

var _ = Context("Resource Workflows", Serial, func() {
	workflowFactory := RunScheduleWorkflowFactory(config.Configuration{
		DefaultExperiment:      "Default",
		DefaultProvider:        "stub",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newRunSchedule = func() *pipelinesv1.RunSchedule {
		return withIntegrationTestFields(pipelinesv1.RandomRunSchedule())
	}

	DescribeTable("RunSchedule Workflows", AssertWorkflow[*pipelinesv1.RunSchedule],
		Entry("Creation",
			newRunSchedule,
			StubWithIdAndError[*pipelinesv1.RunSchedule],
			workflowFactory.ConstructCreationWorkflow,
		), Entry("Update",
			newRunSchedule,
			StubWithIdAndError[*pipelinesv1.RunSchedule],
			workflowFactory.ConstructUpdateWorkflow,
		), Entry("Deletion succeeds",
			newRunSchedule,
			StubWithEmpty[*pipelinesv1.RunSchedule],
			workflowFactory.ConstructDeletionWorkflow,
		), Entry("Deletion fails",
			newRunSchedule,
			StubWithExistingIdAndError[*pipelinesv1.RunSchedule],
			workflowFactory.ConstructDeletionWorkflow,
		),
	)
})
