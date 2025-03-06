//go:build integration

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	testutil "github.com/sky-uk/kfp-operator/common/testutil/provider"
)

var _ = Context("RunSchedule Resource Workflows", Serial, func() {
	workflowFactory := RunScheduleWorkflowFactory(config.KfpControllerConfigSpec{
		DefaultProvider: "not-used",
		DefaultProviderValues: config.DefaultProviderValues{
			ServicePort: 8080,
		},
		DefaultExperiment:      "Default",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newRunSchedule = func() *pipelineshub.RunSchedule {
		return withIntegrationTestFields(pipelineshub.RandomRunSchedule(TestProvider))
	}

	newRunScheduleWithProviderId := func(providerId string) *pipelineshub.RunSchedule {
		rsd := newRunSchedule()
		rsd.SetStatus(
			pipelineshub.Status{
				Provider: pipelineshub.ProviderAndId{
					Id: providerId,
				},
			},
		)

		return rsd
	}

	DescribeTable("Workflows", AssertWorkflow[*pipelineshub.RunSchedule],
		Entry("Creation",
			newRunSchedule(),
			base.Output{Id: testutil.CreateRunScheduleSucceeded},
			workflowFactory.ConstructCreationWorkflow,
		), Entry("Update",
			newRunSchedule(),
			base.Output{Id: testutil.UpdateRunScheduleSucceeded},
			workflowFactory.ConstructUpdateWorkflow,
		), Entry("Deletion succeeds",
			newRunSchedule(),
			base.Output{},
			workflowFactory.ConstructDeletionWorkflow,
		), Entry("Deletion fails",
			newRunScheduleWithProviderId(testutil.DeleteRunScheduledFail),
			base.Output{
				Id:            testutil.DeleteRunScheduledFail,
				ProviderError: (&testutil.DeleteRunScheduleError{}).Error(),
			},
			workflowFactory.ConstructDeletionWorkflow,
		),
	)
})
