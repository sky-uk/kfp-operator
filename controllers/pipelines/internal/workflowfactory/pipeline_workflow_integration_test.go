//go:build integration

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/internal/config"
	testutil "github.com/sky-uk/kfp-operator/pkg/common/testutil/provider"
	"github.com/sky-uk/kfp-operator/pkg/providers/base"
)

var _ = Context("Pipeline Resource Workflows", Serial, func() {
	workflowFactory := PipelineWorkflowFactory(
		config.ConfigSpec{
			DefaultProvider: "not-used",
			DefaultProviderValues: config.DefaultProviderValues{
				ServicePort: 8080,
			},
			DefaultExperiment:      "Default",
			WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
			WorkflowNamespace:      "argo",
		},
	)

	var newPipeline = func() *pipelineshub.Pipeline {
		pipeline := withIntegrationTestFields(pipelineshub.RandomPipeline(TestProvider))
		pipeline.Spec.Image = "kfp-operator-stub-compiler"
		pipeline.Spec.Framework.Name = "stub"

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
