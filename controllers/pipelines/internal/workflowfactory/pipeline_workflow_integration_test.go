//go:build integration

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
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

	var newPipeline = func() *pipelinesv1.Pipeline {
		pipeline := withIntegrationTestFields(pipelinesv1.RandomPipeline(TestProvider))
		pipeline.Spec.Image = "kfp-operator-stub-provider"

		return pipeline
	}

	newPipelineWithProviderId := func(providerId string) *pipelinesv1.Pipeline {
		pipeline := newPipeline()
		pipeline.SetStatus(
			pipelinesv1.Status{
				Provider: pipelinesv1.ProviderAndId{
					Id: providerId,
				},
			},
		)

		return pipeline
	}

	DescribeTable("Pipeline Workflows", AssertWorkflow[*pipelinesv1.Pipeline],
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
