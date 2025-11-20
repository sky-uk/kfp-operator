//go:build integration

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/internal/config"
	testutil "github.com/sky-uk/kfp-operator/pkg/common/testutil/provider"
	"github.com/sky-uk/kfp-operator/pkg/providers/base"
)

var _ = Context("Experiment Resource Workflows", Serial, func() {
	workflowFactory := ExperimentWorkflowFactory(config.ConfigSpec{
		DefaultProvider: "not-used",
		DefaultProviderValues: config.DefaultProviderValues{
			ServicePort: 8080,
		},
		DefaultExperiment:      "Default",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newExperiment = func() *pipelineshub.Experiment {
		return withIntegrationTestFields(pipelineshub.RandomExperiment(TestProvider))
	}

	newExperimentWithProviderId := func(providerId string) *pipelineshub.Experiment {
		experiment := newExperiment()
		experiment.SetStatus(
			pipelineshub.Status{
				Provider: pipelineshub.ProviderAndId{
					Id: providerId,
				},
			},
		)

		return experiment
	}

	DescribeTable("Workflows", AssertWorkflow[*pipelineshub.Experiment],
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
