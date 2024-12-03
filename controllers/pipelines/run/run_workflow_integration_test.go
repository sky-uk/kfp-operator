//go:build integration

package run

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Context("Resource Workflows", Serial, func() {
	workflowFactory := RunWorkflowFactory(config.KfpControllerConfigSpec{
		DefaultExperiment:      "Default",
		DefaultProvider:        "not-used",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newRun = func() *pipelinesv1.Run {
		return testutil.WithIntegrationTestFields(pipelinesv1.RandomRun(testutil.TestProvider))
	}

	DescribeTable("Run Workflows", testutil.AssertWorkflow[*pipelinesv1.Run],
		Entry("Creation",
			newRun,
			testutil.StubWithIdAndError[*pipelinesv1.Run],
			workflowFactory.ConstructCreationWorkflow,
		), Entry("Deletion succeeds",
			newRun,
			testutil.StubWithEmpty[*pipelinesv1.Run],
			workflowFactory.ConstructDeletionWorkflow,
		), Entry("Deletion fails",
			newRun,
			testutil.StubWithExistingIdAndError[*pipelinesv1.Run],
			workflowFactory.ConstructDeletionWorkflow,
		),
	)

	Describe("Update fails", func() {
		It("fails the workflow", func() {
			testCtx := testutil.WorkflowTestHelper[*pipelinesv1.Run]{
				Resource: newRun(),
			}

			workflow, err := workflowFactory.ConstructUpdateWorkflow(*testutil.TestProviderConfig, testCtx.Resource)

			Expect(err).NotTo(HaveOccurred())

			testutil.StubProvider(base.Output{}, testCtx.Resource)
			Expect(testutil.K8sClient.Create(testutil.Ctx, workflow)).To(Succeed())

			Eventually(testCtx.WorkflowByNameToMatch(types.NamespacedName{Name: workflow.Name, Namespace: workflow.Namespace},
				func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
				}), testutil.TestTimeout).Should(Succeed())
		})
	})
})
