//go:build integration
// +build integration

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Context("Resource Workflows", Serial, func() {
	workflowFactory := RunWorkflowFactory(config.Configuration{
		DefaultExperiment:      "Default",
		DefaultProvider:        "stub",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newRun = func() *pipelinesv1.Run {
		return withIntegrationTestFields(pipelinesv1.RandomRun())
	}

	DescribeTable("Run Workflows", AssertWorkflow[*pipelinesv1.Run],
		Entry("Creation",
			newRun,
			StubWithIdAndError[*pipelinesv1.Run],
			workflowFactory.ConstructCreationWorkflow,
		), Entry("Deletion succeeds",
			newRun,
			StubWithEmpty[*pipelinesv1.Run],
			workflowFactory.ConstructDeletionWorkflow,
		), Entry("Deletion fails",
			newRun,
			StubWithExistingIdAndError[*pipelinesv1.Run],
			workflowFactory.ConstructDeletionWorkflow,
		),
	)

	Describe("Update fails", func() {
		It("fails the workflow", func() {
			testCtx := WorkflowTestHelper[*pipelinesv1.Run]{
				Resource: newRun(),
			}

			workflow, err := workflowFactory.ConstructUpdateWorkflow("stub", testCtx.Resource)

			Expect(err).NotTo(HaveOccurred())

			StubProvider(base.Output{}, testCtx.Resource)
			Expect(k8sClient.Create(ctx, workflow)).To(Succeed())

			Eventually(testCtx.WorkflowByNameToMatch(types.NamespacedName{Name: workflow.Name, Namespace: workflow.Namespace},
				func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
				}), TestTimeout).Should(Succeed())
		})
	})
})
