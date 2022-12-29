//go:build integration
// +build integration

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Context("Resource Workflows", Serial, func() {
	workflowFactory := ExperimentWorkflowFactory(config.Configuration{
		DefaultExperiment:      "Default",
		DefaultProvider:        "stub",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newExperiment = func() *pipelinesv1.Experiment {
		return &pipelinesv1.Experiment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      apis.RandomLowercaseString(),
				Namespace: "argo",
			},
			Status: pipelinesv1.Status{
				ProviderId: pipelinesv1.ProviderAndId{
					Provider: "stub",
					Id:       apis.RandomString(),
				},
			},
		}
	}

	DescribeTable("Creation Workflow", AssertWorkflow[*pipelinesv1.Experiment],
		Entry("Creation succeeds",
			newExperiment,
			SucceedCreation[*pipelinesv1.Experiment],
			workflowFactory.ConstructCreationWorkflow,
		), Entry("Creation fails",
			newExperiment,
			FailCreation[*pipelinesv1.Experiment],
			workflowFactory.ConstructCreationWorkflow,
		), Entry("Updating succeeds",
			newExperiment,
			SucceedUpdating[*pipelinesv1.Experiment],
			workflowFactory.ConstructUpdateWorkflow,
		), Entry("Updating fails",
			newExperiment,
			FailUpdating[*pipelinesv1.Experiment],
			workflowFactory.ConstructUpdateWorkflow,
		), Entry("Deletion succeeds",
			newExperiment,
			SucceedDeletion[*pipelinesv1.Experiment],
			workflowFactory.ConstructDeletionWorkflow,
		), Entry("Deletion fails",
			newExperiment,
			FailDeletion[*pipelinesv1.Experiment],
			workflowFactory.ConstructDeletionWorkflow,
		),
	)
})
