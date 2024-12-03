//go:build integration

package pipeline

import (
	. "github.com/onsi/ginkgo/v2"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
)

var _ = Context("Resource Workflows", Serial, func() {
	workflowFactory := PipelineWorkflowFactory(config.KfpControllerConfigSpec{
		DefaultExperiment:      "Default",
		DefaultProvider:        "not-used",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newPipeline = func() *pipelinesv1.Pipeline {
		pipeline := testutil.WithIntegrationTestFields(pipelinesv1.RandomPipeline(testutil.TestProvider))
		pipeline.Spec.Image = "kfp-operator-stub-provider"

		return pipeline
	}

	DescribeTable("Pipeline Workflows", testutil.AssertWorkflow[*pipelinesv1.Pipeline],
		Entry("Creation",
			newPipeline,
			testutil.StubWithIdAndError[*pipelinesv1.Pipeline],
			workflowFactory.ConstructCreationWorkflow,
		), Entry("Update",
			newPipeline,
			testutil.StubWithIdAndError[*pipelinesv1.Pipeline],
			workflowFactory.ConstructUpdateWorkflow,
		), Entry("Deletion succeeds",
			newPipeline,
			testutil.StubWithEmpty[*pipelinesv1.Pipeline],
			workflowFactory.ConstructDeletionWorkflow,
		), Entry("Deletion fails",
			newPipeline,
			testutil.StubWithExistingIdAndError[*pipelinesv1.Pipeline],
			workflowFactory.ConstructDeletionWorkflow,
		),
	)
})
