//go:build integration

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
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

	var newRunSchedule = func() *pipelinesv1.RunSchedule {
		return withIntegrationTestFields(pipelinesv1.RandomRunSchedule(TestProvider))
	}

	newRunScheduleWithProviderId := func(providerId string) *pipelinesv1.RunSchedule {
		rsd := newRunSchedule()
		rsd.SetStatus(
			pipelinesv1.Status{
				Provider: pipelinesv1.ProviderAndId{
					Id: providerId,
				},
			},
		)

		return rsd
	}

	DescribeTable("Workflows", AssertWorkflow[*pipelinesv1.RunSchedule],
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
