//go:build integration

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

var _ = Context("Resource Workflows", Serial, func() {
	workflowFactory := ExperimentWorkflowFactory(config.KfpControllerConfigSpec{
		DefaultExperiment:      "Default",
		DefaultProvider:        "not-used",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newExperiment = func() *pipelineshub.Experiment {
		resource := pipelineshub.RandomExperiment(TestProvider)
		resourceStatus := resource.GetStatus()
		resourceStatus.Provider.Name = TestProvider
		resource.SetStatus(resourceStatus)
		return resource
	}

	DescribeTable("Experiment Workflows", AssertWorkflow[*pipelineshub.Experiment],
		Entry("Creation",
			newExperiment,
			StubWithIdAndError[*pipelineshub.Experiment],
			workflowFactory.ConstructCreationWorkflow,
		), Entry("Update",
			newExperiment,
			StubWithIdAndError[*pipelineshub.Experiment],
			workflowFactory.ConstructUpdateWorkflow,
		), Entry("Deletion succeeds",
			newExperiment,
			StubWithEmpty[*pipelineshub.Experiment],
			workflowFactory.ConstructDeletionWorkflow,
		), Entry("Deletion fails",
			newExperiment,
			StubWithExistingIdAndError[*pipelineshub.Experiment],
			workflowFactory.ConstructDeletionWorkflow,
		),
	)
})
