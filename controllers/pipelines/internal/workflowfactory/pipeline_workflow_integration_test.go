//go:build integration

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	testutil "github.com/sky-uk/kfp-operator/common/testutil/provider"
)

var _ = Context("Pipeline Resource Workflows", Serial, func() {
	workflowFactory := PipelineWorkflowFactory(
		config.KfpControllerConfigSpec{
			DefaultProvider: "not-used",
			DefaultProviderValues: config.DefaultProviderValues{
				ServicePort: 8080,
			},
			DefaultExperiment:      "Default",
			WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
			WorkflowNamespace:      "argo",
			PipelineFrameworkImages: map[string]string{
				"default": "kfp-operator-stub-provider",
			},
		},
	)

	var newPipeline = func() *pipelineshub.Pipeline {
		pipeline := withIntegrationTestFields(pipelineshub.RandomPipeline(TestProvider))
		pipeline.Spec.Image = "kfp-operator-stub-provider"

		return pipeline
	}

	newPipelineWithProviderId := func(providerId string) *pipelineshub.Pipeline {
		pipeline := newPipeline()
		pipeline.SetStatus(
			pipelineshub.Status{
				Provider: pipelineshub.ProviderAndId{
					Id: providerId,
				},
			},
		)

		return pipeline
	}

	DescribeTable("Workflows", AssertWorkflow[*pipelineshub.Pipeline],
		Entry(
			"Creation",
			newPipeline(),
			base.Output{Id: testutil.CreatePipelineSucceeded},
			workflowFactory.ConstructCreationWorkflow,
		), Entry(
			"Update",
			newPipeline(),
			base.Output{Id: testutil.UpdatePipelineSucceeded},
			workflowFactory.ConstructUpdateWorkflow,
		), Entry(
			"Deletion succeeds",
			newPipeline(),
			base.Output{},
			workflowFactory.ConstructDeletionWorkflow,
		),
		Entry(
			"Deletion fails",
			newPipelineWithProviderId(testutil.DeletePipelineFail),
			base.Output{
				Id:            testutil.DeletePipelineFail,
				ProviderError: (&testutil.DeletePipelineError{}).Error(),
			},
			workflowFactory.ConstructDeletionWorkflow,
		),
	)
})
