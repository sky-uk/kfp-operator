//go:build integration
// +build integration

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"github.com/walkerus/go-wiremock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Context("RunConfiguration Workflows", func() {
	workflowFactory := RunConfigurationWorkflowFactory{
		Config: configv1.Configuration{
			KfpEndpoint:     "http://wiremock:80",
			KfpSdkImage:     "kfp-operator-argo-kfp-sdk",
			ImagePullPolicy: "Never", // Needed for minikube to use local images
			DefaultExperiment: "Default",
		},
	}

	var KfpCreateToSucceed = func(rcsName string, kfpId string) error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/jobs")).
			WillReturn(
				`{"id": "`+kfpId+`", "created_at": "2021-09-10T15:46:08Z", "name": "`+rcsName+`"}`,
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var KfpCreateToFail = func(rcsName string, kfpId string) error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/jobs")).
			WillReturn(
				`{"status": "failed"}`,
				map[string]string{"Content-Type": "application/json"},
				404,
			))
	}

	var KfpGetExperimentToSucceed = func(experimentName string, experimentId string) error {
		return wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/experiments")).
			WithQueryParam("filter", wiremock.EqualTo(`{"predicates": [{"op": 1, "key": "name", "stringValue": "`+experimentName+`"}]}`)).
			WillReturn(
				`{"experiments": [{"id": "`+experimentId+`", "created_at": "2021-09-10T15:46:08Z", "name": "`+experimentName+`"}]}`,
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var KfpGetPipelineToSucceed = func(pipelineName string, pipelineId string) error {
		return wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines")).
			WithQueryParam("filter", wiremock.EqualTo(`{"predicates": [{"op": 1, "key": "name", "stringValue": "`+pipelineName+`"}]}`)).
			WillReturn(
				`{"experiments": [{"id": "`+pipelineId+`", "created_at": "2021-09-10T15:46:08Z", "name": "`+pipelineName+`"}]}`,
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	Describe("Creation workflow", func() {
		When("The creation succeeds", func() {
			It("Succeeds the workflow with a KfpId", func() {

				testCtx := NewRunconfigurationTestContext(
					&pipelinesv1.RunConfiguration{
						ObjectMeta: metav1.ObjectMeta{
							Name:      RandomLowercaseString(),
							Namespace: "argo",
						},
						Spec: pipelinesv1.RunConfigurationSpec{
							PipelineName: "pipeline",
							Schedule: "* * * * * *",
						},
					},
					k8sClient, ctx)

				Expect(KfpGetExperimentToSucceed(workflowFactory.Config.DefaultExperiment, ExperimentId)).To(Succeed())
				Expect(KfpGetPipelineToSucceed(testCtx.RunConfiguration.Spec.PipelineName, PipelineId)).To(Succeed())
				Expect(KfpCreateToSucceed(testCtx.RunConfiguration.Name, JobId)).To(Succeed())

				workflow, err := workflowFactory.ConstructCreationWorkflow(testCtx.RunConfiguration)
				Expect(err).NotTo(HaveOccurred())

				err = k8sClient.Create(ctx, workflow)
				Expect(err).NotTo(HaveOccurred())

				Eventually(testCtx.WorkflowToMatch(RunConfigurationWorkflowConstants.CreateOperationLabel, func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
					g.Expect(getWorkflowOutput(workflow, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)).
						To(Equal(JobId))
				}), TestTimeout).Should(Succeed())
			})
		})

		When("The creation fails", func() {
			It("Fails the workflow", func() {

				testCtx := NewRunconfigurationTestContext(
					&pipelinesv1.RunConfiguration{
						ObjectMeta: metav1.ObjectMeta{
							Name:      RandomLowercaseString(),
							Namespace: "argo",
						},
						Spec: pipelinesv1.RunConfigurationSpec{
							PipelineName: "pipeline",
							Schedule: "* * * * * *",
						},
					},
					k8sClient, ctx)

				Expect(KfpGetExperimentToSucceed(workflowFactory.Config.DefaultExperiment, ExperimentId)).To(Succeed())
				Expect(KfpGetPipelineToSucceed(testCtx.RunConfiguration.Spec.PipelineName, PipelineId)).To(Succeed())
				Expect(KfpCreateToFail(testCtx.RunConfiguration.Name, JobId)).To(Succeed())

				workflow, err := workflowFactory.ConstructCreationWorkflow(testCtx.RunConfiguration)
				Expect(err).NotTo(HaveOccurred())

				err = k8sClient.Create(ctx, workflow)
				Expect(err).NotTo(HaveOccurred())

				Eventually(testCtx.WorkflowToMatch(RunConfigurationWorkflowConstants.CreateOperationLabel, func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
				}), TestTimeout).Should(Succeed())
			})
		})
	})
})