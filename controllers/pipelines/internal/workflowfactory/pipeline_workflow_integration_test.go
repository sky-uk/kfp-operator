//go:build integration

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	testutil "github.com/sky-uk/kfp-operator/common/testutil/provider"
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

	var newPipeline = func() *pipelinesv1.Pipeline {
		pipeline := withIntegrationTestFields(pipelinesv1.RandomPipeline(TestProvider))
		pipeline.Spec.Image = "kfp-operator-stub-provider"

		return pipeline
	}

	DescribeTable("Pipeline Workflows", AssertWorkflow[*pipelinesv1.Pipeline],
		Entry(
			"Creation",
			newPipeline,
			base.Output{Id: testutil.CreatePipelineSucceeded},
			workflowFactory.ConstructCreationWorkflow,
		),
		// ), Entry(
		// 	"Update",
		// 	newPipeline,
		// 	StubWithIdAndError[*pipelinesv1.Pipeline],
		// 	workflowFactory.ConstructUpdateWorkflow,
		// ), Entry(
		// 	"Deletion succeeds",
		// 	newPipeline,
		// 	StubWithEmpty[*pipelinesv1.Pipeline],
		// 	workflowFactory.ConstructDeletionWorkflow,
		// ), Entry(
		// 	"Deletion fails",
		// 	newPipeline,
		// 	StubWithExistingIdAndError[*pipelinesv1.Pipeline],
		// 	workflowFactory.ConstructDeletionWorkflow,
		// ),
	)
})
