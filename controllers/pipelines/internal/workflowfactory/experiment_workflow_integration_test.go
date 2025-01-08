//go:build integration

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	testutil "github.com/sky-uk/kfp-operator/common/testutil/provider"
)

var _ = Context("Experiment Resource Workflows", Serial, func() {
	workflowFactory := ExperimentWorkflowFactory(config.KfpControllerConfigSpec{
		DefaultProvider: "not-used",
		DefaultProviderValues: config.DefaultProviderValues{
			ServicePort: 8080,
		},
		DefaultExperiment:      "Default",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newExperiment = func() *pipelinesv1.Experiment {
		return withIntegrationTestFields(pipelinesv1.RandomExperiment(TestProvider.Name))
	}

	newExperimentWithProviderId := func(providerId string) *pipelinesv1.Experiment {
		experiment := newExperiment()
		experiment.SetStatus(
			pipelinesv1.Status{
				Provider: pipelinesv1.ProviderAndId{
					Id: providerId,
				},
			},
		)

		return experiment
	}

	DescribeTable("Workflows", AssertWorkflow[*pipelinesv1.Experiment],
		Entry(
			"Creation",
			newExperiment(),
			base.Output{Id: testutil.CreateExperimentSucceeded},
			workflowFactory.ConstructCreationWorkflow,
		), Entry(
			"Update",
			newExperiment(),
			base.Output{Id: testutil.UpdateExperimentSucceeded},
			workflowFactory.ConstructUpdateWorkflow,
		), Entry(
			"Deletion succeeds",
			newExperiment(),
			base.Output{},
			workflowFactory.ConstructDeletionWorkflow,
		), Entry(
			"Deletion fails",
			newExperimentWithProviderId(testutil.DeleteExperimentFail),
			base.Output{
				Id:            testutil.DeleteExperimentFail,
				ProviderError: (&testutil.DeleteExperimentError{}).Error(),
			},
			workflowFactory.ConstructDeletionWorkflow,
		),
	)
})
