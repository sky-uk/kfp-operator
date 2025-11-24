//go:build integration

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/internal/config"
	testutil "github.com/sky-uk/kfp-operator/pkg/common/testutil/provider"
	"github.com/sky-uk/kfp-operator/pkg/providers/base"
)

var _ = Context("RunSchedule Resource Workflows", Serial, func() {
	workflowFactory := RunScheduleWorkflowFactory(config.ConfigSpec{
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
