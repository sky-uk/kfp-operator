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

var _ = Context("Pipeline Workflows", func() {

	workflowFactory := PipelineWorkflowFactory{
		WorkflowFactory: WorkflowFactory{
			Config: configv1.Configuration{
				KfpEndpoint:     "http://wiremock:80",
				KfpSdkImage:     "kfp-operator-argo-kfp-sdk",
				CompilerImage:   "kfp-operator-argo-compiler",
				ImagePullPolicy: "Never", // Needed for minikube to use local images
				PipelineStorage: "gs://some-bucket",
				DefaultBeamArgs: map[string]string{
					"project": "project",
				},
			},
		},
	}

	var SucceedUpload = func(pipelineName string, pipelineId string) error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload")).
			WithQueryParam("name", wiremock.EqualTo(pipelineName)).
			WillReturn(
				`{"id": "`+pipelineId+`", "created_at": "2021-09-10T15:46:08Z", "name": "`+pipelineName+`"}`,
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var FailUpload = func(pipelineName string, pipelineId string) error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload")).
			WithQueryParam("name", wiremock.EqualTo(pipelineName)).
			WillReturn(
				`{"status": "failed"}`,
				map[string]string{"Content-Type": "application/json"},
				404,
			))
	}

	var SucceedUploadVersion = func(pipelineName string, pipelineId string, pipelineVersion string) error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload_version")).
			WithQueryParam("name", wiremock.EqualTo(pipelineVersion)).
			WithQueryParam("pipelineid", wiremock.EqualTo(pipelineId)).
			WillReturn(
				`{"id": "`+pipelineVersion+`", "created_at": "2021-09-10T15:46:08Z", "name": "pipeline", "resource_references": [{"key": {"id": "`+pipelineId+`", "apiResourceType": "PIPELINE"}, "name": "`+pipelineName+`", "relationship": "OWNER"}]}`,
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var FailUploadVersion = func(pipelineId string, pipelineVersion string) error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload_version")).
			WithQueryParam("name", wiremock.EqualTo(pipelineVersion)).
			WithQueryParam("pipelineid", wiremock.EqualTo(pipelineId)).
			WillReturn(
				`{"status": "failed"`,
				map[string]string{"Content-Type": "application/json"},
				400,
			))
	}

	var SucceedDeletion = func(pipelineId string) error {
		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/"+pipelineId)).
			WillReturn(
				`{"status": "deleted"}`,
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var FailDeletion = func(pipelineId string) error {
		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/"+pipelineId)).
			WillReturn(
				`{"status": "failed"}`,
				map[string]string{"Content-Type": "application/json"},
				400,
			))
	}

	Describe("Creation workflow", func() {
		When("The creation and update succeeds", func() {
			It("Succeeds the workflow with a KfpId", func() {

				testCtx := NewPipelineTestContext(
					&pipelinesv1.Pipeline{
						ObjectMeta: metav1.ObjectMeta{
							Name:      RandomLowercaseString(),
							Namespace: "argo",
						},
						Spec: pipelineSpec,
					},
					k8sClient, ctx)

				Expect(SucceedUpload(testCtx.Pipeline.Name, PipelineId)).To(Succeed())
				Expect(SucceedUploadVersion(testCtx.Pipeline.Name, PipelineId, V1)).To(Succeed())

				workflow, err := workflowFactory.ConstructCreationWorkflow(testCtx.Pipeline.Spec, testCtx.Pipeline.ObjectMeta, V1)
				Expect(err).NotTo(HaveOccurred())

				err = k8sClient.Create(ctx, workflow)
				Expect(err).NotTo(HaveOccurred())

				Eventually(testCtx.WorkflowByOperationToMatch(PipelineWorkflowConstants.CreateOperationLabel, func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
					g.Expect(getWorkflowOutput(workflow, PipelineWorkflowConstants.PipelineIdParameterName)).
						To(Equal(PipelineId))
				}), TestTimeout).Should(Succeed())
			})
		})

		When("The creation succeeds but the update fails", func() {
			It("Fails the workflow", func() {
				testCtx := NewPipelineTestContext(
					&pipelinesv1.Pipeline{
						ObjectMeta: metav1.ObjectMeta{
							Name:      RandomLowercaseString(),
							Namespace: "argo",
						},
						Spec: pipelineSpec,
					},
					k8sClient, ctx)

				Expect(FailUpload(testCtx.Pipeline.Name, PipelineId)).To(Succeed())
				Expect(FailUploadVersion(testCtx.Pipeline.Name, PipelineId)).To(Succeed())

				workflow, err := workflowFactory.ConstructCreationWorkflow(testCtx.Pipeline.Spec, testCtx.Pipeline.ObjectMeta, V1)
				Expect(err).NotTo(HaveOccurred())

				err = k8sClient.Create(ctx, workflow)
				Expect(err).NotTo(HaveOccurred())

				Eventually(testCtx.WorkflowByOperationToMatch(PipelineWorkflowConstants.CreateOperationLabel, func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
				}), TestTimeout).Should(Succeed())
			})
		})

		When("The creation fails", func() {
			It("Fails the workflow", func() {
				testCtx := NewPipelineTestContext(
					&pipelinesv1.Pipeline{
						ObjectMeta: metav1.ObjectMeta{
							Name:      RandomLowercaseString(),
							Namespace: "argo",
						},
						Spec: pipelineSpec,
					},
					k8sClient, ctx)

				Expect(SucceedUpload(testCtx.Pipeline.Name, PipelineId)).To(Succeed())
				Expect(FailUploadVersion(PipelineId, V1)).To(Succeed())

				workflow, err := workflowFactory.ConstructCreationWorkflow(testCtx.Pipeline.Spec, testCtx.Pipeline.ObjectMeta, V1)
				Expect(err).NotTo(HaveOccurred())

				err = k8sClient.Create(ctx, workflow)
				Expect(err).NotTo(HaveOccurred())

				Eventually(testCtx.WorkflowByOperationToMatch(PipelineWorkflowConstants.CreateOperationLabel, func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
				}), TestTimeout).Should(Succeed())
			})
		})
	})

	Describe("Update workflow", func() {
		When("The upload succeeds", func() {
			It("Succeeds the workflow", func() {
				testCtx := NewPipelineTestContext(
					&pipelinesv1.Pipeline{
						ObjectMeta: metav1.ObjectMeta{
							Name:      RandomLowercaseString(),
							Namespace: "argo",
						},
						Spec: pipelineSpec,
					},
					k8sClient, ctx)

				Expect(SucceedUploadVersion(testCtx.Pipeline.Name, PipelineId, V1)).To(Succeed())

				workflow, err := workflowFactory.ConstructUpdateWorkflow(testCtx.Pipeline.Spec, testCtx.Pipeline.ObjectMeta, PipelineId, V1)
				Expect(err).NotTo(HaveOccurred())

				err = k8sClient.Create(ctx, workflow)
				Expect(err).NotTo(HaveOccurred())

				Eventually(testCtx.WorkflowByOperationToMatch(PipelineWorkflowConstants.UpdateOperationLabel, func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				}), TestTimeout).Should(Succeed())
			})
		})

		When("The upload fails", func() {
			It("Fails the workflow", func() {
				testCtx := NewPipelineTestContext(
					&pipelinesv1.Pipeline{
						ObjectMeta: metav1.ObjectMeta{
							Name:      RandomLowercaseString(),
							Namespace: "argo",
						},
						Spec: pipelineSpec,
					},
					k8sClient, ctx)

				Expect(FailUploadVersion(PipelineId, V1)).To(Succeed())

				workflow, err := workflowFactory.ConstructUpdateWorkflow(testCtx.Pipeline.Spec, testCtx.Pipeline.ObjectMeta, PipelineId, V1)
				Expect(err).NotTo(HaveOccurred())

				err = k8sClient.Create(ctx, workflow)
				Expect(err).NotTo(HaveOccurred())

				Eventually(testCtx.WorkflowByOperationToMatch(PipelineWorkflowConstants.UpdateOperationLabel, func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
				}), TestTimeout).Should(Succeed())
			})
		})
	})

	Describe("Deletion workflow", func() {
		When("The deletion succeeds", func() {
			It("Succeeds the workflow", func() {
				testCtx := NewPipelineTestContext(
					&pipelinesv1.Pipeline{
						ObjectMeta: metav1.ObjectMeta{
							Name:      RandomLowercaseString(),
							Namespace: "argo",
						},
						Spec: pipelineSpec,
						Status: pipelinesv1.PipelineStatus{
							KfpId: PipelineId,
						},
					},
					k8sClient, ctx)

				Expect(SucceedDeletion(PipelineId)).To(Succeed())

				workflow := workflowFactory.ConstructDeletionWorkflow(testCtx.Pipeline.ObjectMeta, PipelineId)

				err := k8sClient.Create(ctx, workflow)
				Expect(err).NotTo(HaveOccurred())

				Eventually(testCtx.WorkflowByOperationToMatch(PipelineWorkflowConstants.DeleteOperationLabel, func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				}), TestTimeout).Should(Succeed())
			})
		})

		When("The deletion fails", func() {
			It("Fails the workflow", func() {
				testCtx := NewPipelineTestContext(
					&pipelinesv1.Pipeline{
						ObjectMeta: metav1.ObjectMeta{
							Name:      RandomLowercaseString(),
							Namespace: "argo",
						},
						Spec: pipelineSpec,
						Status: pipelinesv1.PipelineStatus{
							KfpId: PipelineId,
						},
					},
					k8sClient, ctx)

				Expect(FailDeletion(PipelineId)).To(Succeed())

				workflow := workflowFactory.ConstructDeletionWorkflow(testCtx.Pipeline.ObjectMeta, PipelineId)

				err := k8sClient.Create(ctx, workflow)
				Expect(err).NotTo(HaveOccurred())

				Eventually(testCtx.WorkflowByOperationToMatch(PipelineWorkflowConstants.DeleteOperationLabel, func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
				}), TestTimeout).Should(Succeed())
			})
		})
	})
})
