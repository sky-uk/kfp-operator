//go:build integration

package workflowfactory

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	. "github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Context("Resource Workflows", Serial, func() {
	workflowFactory := RunWorkflowFactory(config.KfpControllerConfigSpec{
		DefaultExperiment:      "Default",
		DefaultProvider:        "not-used",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newRun = func() *pipelineshub.Run {
		return withIntegrationTestFields(pipelineshub.RandomRun(TestProvider))
	}

	DescribeTable("Run Workflows", AssertWorkflow[*pipelineshub.Run],
		Entry("Creation",
			newRun,
			StubWithIdAndError[*pipelineshub.Run],
			workflowFactory.ConstructCreationWorkflow,
		), Entry("Deletion succeeds",
			newRun,
			StubWithEmpty[*pipelineshub.Run],
			workflowFactory.ConstructDeletionWorkflow,
		), Entry("Deletion fails",
			newRun,
			StubWithExistingIdAndError[*pipelineshub.Run],
			workflowFactory.ConstructDeletionWorkflow,
		),
	)

	Describe("Update fails", func() {
		It("fails the workflow", func() {
			testCtx := WorkflowTestHelper[*pipelineshub.Run]{
				Resource: newRun(),
			}

			workflow, err := workflowFactory.ConstructUpdateWorkflow(*TestProviderConfig, testCtx.Resource)

			Expect(err).NotTo(HaveOccurred())

			StubProvider(base.Output{}, testCtx.Resource)
			Expect(K8sClient.Create(Ctx, workflow)).To(Succeed())

			Eventually(
				testCtx.WorkflowByNameToMatch(
					types.NamespacedName{
						Name:      workflow.Name,
						Namespace: workflow.Namespace,
					},
					func(g Gomega, workflow *argo.Workflow) {
						g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
					},
				),
				TestTimeout,
			).Should(Succeed())
		})
	})
})
