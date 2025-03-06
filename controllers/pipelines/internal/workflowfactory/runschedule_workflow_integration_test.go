//go:build integration

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

var _ = Context("Resource Workflows", Serial, func() {
	workflowFactory := RunScheduleWorkflowFactory(config.KfpControllerConfigSpec{
		DefaultExperiment:      "Default",
		DefaultProvider:        "not-used",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newRunSchedule = func() *pipelineshub.RunSchedule {
		return withIntegrationTestFields(pipelineshub.RandomRunSchedule(TestProvider))
	}

	DescribeTable("RunSchedule Workflows", AssertWorkflow[*pipelineshub.RunSchedule],
		Entry("Creation",
			newRunSchedule,
			StubWithIdAndError[*pipelineshub.RunSchedule],
			workflowFactory.ConstructCreationWorkflow,
		), Entry("Update",
			newRunSchedule,
			StubWithIdAndError[*pipelineshub.RunSchedule],
			workflowFactory.ConstructUpdateWorkflow,
		), Entry("Deletion succeeds",
			newRunSchedule,
			StubWithEmpty[*pipelineshub.RunSchedule],
			workflowFactory.ConstructDeletionWorkflow,
		), Entry("Deletion fails",
			newRunSchedule,
			StubWithExistingIdAndError[*pipelineshub.RunSchedule],
			workflowFactory.ConstructDeletionWorkflow,
		),
	)
})
