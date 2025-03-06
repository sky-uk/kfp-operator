//go:build integration

package workflowfactory

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	testutil "github.com/sky-uk/kfp-operator/common/testutil/provider"
	. "github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Context("Run Resource Workflows", Serial, func() {
	workflowFactory := RunWorkflowFactory(config.KfpControllerConfigSpec{
		DefaultProvider: "not-used",
		DefaultProviderValues: config.DefaultProviderValues{
			ServicePort: 8080,
		},
		DefaultExperiment:      "Default",
		WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
		WorkflowNamespace:      "argo",
	})

	var newRun = func() *pipelineshub.Run {
		return withIntegrationTestFields(pipelineshub.RandomRun(TestProvider))
	}

	newRunWithProviderId := func(providerId string) *pipelineshub.Run {
		run := newRun()
		run.SetStatus(
			pipelineshub.Status{
				Provider: pipelineshub.ProviderAndId{
					Id: providerId,
				},
			},
		)

		return run
	}

	DescribeTable("Workflows", AssertWorkflow[*pipelineshub.Run],
		Entry(
			"Creation",
			newRun(),
			base.Output{Id: testutil.CreateRunSucceeded},
			workflowFactory.ConstructCreationWorkflow,
		), Entry(
			"Deletion succeeds",
			newRun(),
			base.Output{},
			workflowFactory.ConstructDeletionWorkflow,
		), Entry(
			"Deletion fails",
			newRunWithProviderId(testutil.DeleteRunFail),
			base.Output{
				Id:            testutil.DeleteRunFail,
				ProviderError: (&testutil.DeleteRunError{}).Error(),
			},
			workflowFactory.ConstructDeletionWorkflow,
		),
	)

	Describe("Update fails", func() {
		It("fails the workflow", func() {
			testCtx := WorkflowTestHelper[*pipelineshub.Run]{
				Resource: newRun(),
			}

			providerSvc := corev1.Service{}
			err := K8sClient.Get(
				Ctx,
				types.NamespacedName{
					Namespace: TestNamespace,
					Name:      "provider-test",
				},
				&providerSvc,
			)
			Expect(err).ToNot(HaveOccurred())

			workflow, err := workflowFactory.ConstructUpdateWorkflow(
				*TestProviderConfig,
				providerSvc,
				testCtx.Resource,
			)
			Expect(err).NotTo(HaveOccurred())

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
