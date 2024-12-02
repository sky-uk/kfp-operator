//go:build integration

package runschedule

import (
	. "github.com/onsi/ginkgo/v2"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
)

var _ = Context("Resource Workflows", Serial, func() {
	workflowFactory := RunScheduleWorkflowFactory(config.KfpControllerConfigSpec{
		DefaultExperiment:      "Default",
		DefaultProvider:        "not-used",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newRunSchedule = func() *pipelinesv1.RunSchedule {
		return testutil.WithIntegrationTestFields(pipelinesv1.RandomRunSchedule(testutil.TestProvider))
	}

	DescribeTable("RunSchedule Workflows", testutil.AssertWorkflow[*pipelinesv1.RunSchedule],
		Entry("Creation",
			newRunSchedule,
			testutil.StubWithIdAndError[*pipelinesv1.RunSchedule],
			workflowFactory.ConstructCreationWorkflow,
		), Entry("Update",
			newRunSchedule,
			testutil.StubWithIdAndError[*pipelinesv1.RunSchedule],
			workflowFactory.ConstructUpdateWorkflow,
		), Entry("Deletion succeeds",
			newRunSchedule,
			testutil.StubWithEmpty[*pipelinesv1.RunSchedule],
			workflowFactory.ConstructDeletionWorkflow,
		), Entry("Deletion fails",
			newRunSchedule,
			testutil.StubWithExistingIdAndError[*pipelinesv1.RunSchedule],
			workflowFactory.ConstructDeletionWorkflow,
		),
	)
})
