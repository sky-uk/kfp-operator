//go:build integration

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

var _ = Context("Resource Workflows", Serial, func() {
	workflowFactory := PipelineWorkflowFactory(config.KfpControllerConfigSpec{
		DefaultExperiment:      "Default",
		DefaultProvider:        "not-used",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
		PipelineFrameworkImages: map[string]string{
			"default": "kfp-operator-stub-provider",
		},
	})

	var newPipeline = func() *pipelineshub.Pipeline {
		pipeline := withIntegrationTestFields(pipelineshub.RandomPipeline(TestProvider))
		pipeline.Spec.Image = "kfp-operator-stub-provider"

		return pipeline
	}

	DescribeTable("Pipeline Workflows", AssertWorkflow[*pipelineshub.Pipeline],
		Entry("Creation",
			newPipeline,
			StubWithIdAndError[*pipelineshub.Pipeline],
			workflowFactory.ConstructCreationWorkflow,
		), Entry("Update",
			newPipeline,
			StubWithIdAndError[*pipelineshub.Pipeline],
			workflowFactory.ConstructUpdateWorkflow,
		), Entry("Deletion succeeds",
			newPipeline,
			StubWithEmpty[*pipelineshub.Pipeline],
			workflowFactory.ConstructDeletionWorkflow,
		), Entry("Deletion fails",
			newPipeline,
			StubWithExistingIdAndError[*pipelineshub.Pipeline],
			workflowFactory.ConstructDeletionWorkflow,
		),
	)
})
