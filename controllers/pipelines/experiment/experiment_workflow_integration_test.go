//go:build integration

package experiment

import (
	. "github.com/onsi/ginkgo/v2"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	testutil "github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
)

var _ = Context("Resource Workflows", Serial, func() {
	workflowFactory := ExperimentWorkflowFactory(config.KfpControllerConfigSpec{
		DefaultExperiment:      "Default",
		DefaultProvider:        "not-used",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newExperiment = func() *pipelinesv1.Experiment {
		resource := pipelinesv1.RandomExperiment(testutil.TestProvider)
		resourceStatus := resource.GetStatus()
		resourceStatus.Provider.Name = testutil.TestProvider
		resource.SetStatus(resourceStatus)
		return resource
	}

	DescribeTable("Experiment Workflows", testutil.AssertWorkflow[*pipelinesv1.Experiment],
		Entry("Creation",
			newExperiment,
			testutil.StubWithIdAndError[*pipelinesv1.Experiment],
			workflowFactory.ConstructCreationWorkflow,
		), Entry("Update",
			newExperiment,
			testutil.StubWithIdAndError[*pipelinesv1.Experiment],
			workflowFactory.ConstructUpdateWorkflow,
		), Entry("Deletion succeeds",
			newExperiment,
			testutil.StubWithEmpty[*pipelinesv1.Experiment],
			workflowFactory.ConstructDeletionWorkflow,
		), Entry("Deletion fails",
			newExperiment,
			testutil.StubWithExistingIdAndError[*pipelinesv1.Experiment],
			workflowFactory.ConstructDeletionWorkflow,
		),
	)
})
