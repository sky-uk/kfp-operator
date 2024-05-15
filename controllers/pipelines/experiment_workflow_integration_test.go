//go:build integration

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
)

var _ = Context("Resource Workflows", Serial, func() {
	workflowFactory := ExperimentWorkflowFactory(config.Configuration{
		DefaultExperiment:      "Default",
		DefaultProvider:        "stub",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newExperiment = func() *pipelinesv1.Experiment {
		resource := pipelinesv1.RandomExperiment()
		resourceStatus := resource.GetStatus()
		resourceStatus.ProviderId.Provider = TestProvider
		resource.SetStatus(resourceStatus)
		return resource
	}

	DescribeTable("Experiment Workflows", AssertWorkflow[*pipelinesv1.Experiment],
		Entry("Creation",
			newExperiment,
			StubWithIdAndError[*pipelinesv1.Experiment],
			workflowFactory.ConstructCreationWorkflow,
		), Entry("Update",
			newExperiment,
			StubWithIdAndError[*pipelinesv1.Experiment],
			workflowFactory.ConstructUpdateWorkflow,
		), Entry("Deletion succeeds",
			newExperiment,
			StubWithEmpty[*pipelinesv1.Experiment],
			workflowFactory.ConstructDeletionWorkflow,
		), Entry("Deletion fails",
			newExperiment,
			StubWithExistingIdAndError[*pipelinesv1.Experiment],
			workflowFactory.ConstructDeletionWorkflow,
		),
	)
})
