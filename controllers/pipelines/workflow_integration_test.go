// +build integration

package pipelines

import (
	"context"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"github.com/sky-uk/kfp-operator/external"
	"github.com/walkerus/go-wiremock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Workflows", func() {
	const (
		TestTimeout = 120
	)

	var (
		k8sClient client.Client
		ctx       context.Context

		restCfg = rest.Config{
			Host:    "http://localhost:8080",
			APIPath: "/api",
		}
	
		pipelineSpec = pipelinesv1.PipelineSpec{
			Image:         "kfp-quickstart",
			TfxComponents: "pipeline.create_components",
		}
	
		wiremockClient *wiremock.Client
		workflows      WorkflowFactory
	)

	var KfpUploadToSucceed = func(pipelineName string, pipelineId string) error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload")).
			WithQueryParam("name", wiremock.EqualTo(pipelineName)).
			WillReturn(
				`{"id": "`+pipelineId+`", "created_at": "2021-09-10T15:46:08Z", "name": "`+pipelineName+`"}`,
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}
	
	var KfpUploadToFail = func(pipelineName string, pipelineId string) error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload")).
			WithQueryParam("name", wiremock.EqualTo(pipelineName)).
			WillReturn(
				`{"status": "failed"}`,
				map[string]string{"Content-Type": "application/json"},
				404,
			))
	}
	
	var KfpUploadVersionToReturn = func(pipelineName string, pipelineId string, pipelineVersion string) error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload_version")).
			WithQueryParam("name", wiremock.EqualTo(pipelineVersion)).
			WithQueryParam("pipelineid", wiremock.EqualTo(pipelineId)).
			WillReturn(
				`{"id": "`+pipelineVersion+`", "created_at": "2021-09-10T15:46:08Z", "name": "pipeline", "resource_references": [{"key": {"id": "`+pipelineId+`", "apiResourceType": "PIPELINE"}, "name": "`+pipelineName+`", "relationship": "OWNER"}]}`,
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}
	
	var KfpUploadVersionToFail = func(pipelineId string, pipelineVersion string) error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload_version")).
			WithQueryParam("name", wiremock.EqualTo(pipelineVersion)).
			WithQueryParam("pipelineid", wiremock.EqualTo(pipelineId)).
			WillReturn(
				`{"status": "failed"`,
				map[string]string{"Content-Type": "application/json"},
				400,
			))
	}
	
	var KfpDeleteToReturn = func(pipeline pipelinesv1.Pipeline, pipelineId string) error {
		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/"+pipelineId)).
			WillReturn(
				`{"satus": "deleted"}`,
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}
	
	var KfpDeleteToFail = func(pipeline pipelinesv1.Pipeline, pipelineId string) error {
		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/"+pipelineId)).
			WillReturn(
				`{"satus": "failed"}`,
				map[string]string{"Content-Type": "application/json"},
				400,
			))
	}

	BeforeSuite(func() {
		wiremockClient = wiremock.NewClient("http://localhost:8081")

		Expect(external.InitSchemes(scheme.Scheme)).To(Succeed())
		var err error
		k8sClient, err = client.New(&restCfg, client.Options{Scheme: scheme.Scheme})
		Expect(err).NotTo(HaveOccurred())
		ctx = context.Background()

		workflows = WorkflowFactory{
			Config: configv1.Configuration{
				KfpEndpoint:     "http://wiremock:80",
				KfpToolsImage:   "kfp-tools",
				CompilerImage:   "compiler",
				ImagePullPolicy: "Never", // Needed for minikube to use local images
				PipelineStorage: "gs://some-bucket",
				DataflowProject: "project",
			},
		}
	})

	BeforeEach(func() {
		Expect(wiremockClient.Reset()).To(Succeed())

		Expect(wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/healthz")).
			WillReturn(
				`{"status": "ok"}`,
				map[string]string{"Content-Type": "application/json"},
				200,
			))).To(Succeed())
	})

	Describe("Creation workflow", func() {
		When("The creation and update succeeds", func() {
			It("Succeeds the workflow with a Pipeline Id", func() {

				testCtx := NewTestContextWithPipeline(
					&pipelinesv1.Pipeline{
						ObjectMeta: metav1.ObjectMeta{
							Name:      RandomLowercaseString(),
							Namespace: "argo",
						},
						Spec: pipelineSpec,
					},
					k8sClient, ctx)

				Expect(KfpUploadToSucceed(testCtx.Pipeline.Name, PipelineId)).To(Succeed())
				Expect(KfpUploadVersionToReturn(testCtx.Pipeline.Name, PipelineId, V1)).To(Succeed())

				workflow, err := workflows.ConstructCreationWorkflow(testCtx.Pipeline.Spec, testCtx.Pipeline.ObjectMeta, V1)
				Expect(err).NotTo(HaveOccurred())

				err = k8sClient.Create(ctx, workflow)
				Expect(err).NotTo(HaveOccurred())

				Eventually(testCtx.WorkflowToMatch(CreateOperationLabel, func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
					g.Expect(GetWorkflowOutput(workflow, WorkflowFactoryConstants.pipelineIdParameterName)).
						To(Equal(PipelineId))
				}), TestTimeout).Should(Succeed())
			})
		})

		When("The creation succeeds but the update fails", func() {
			It("Fails the workflow", func() {
				testCtx := NewTestContextWithPipeline(
					&pipelinesv1.Pipeline{
						ObjectMeta: metav1.ObjectMeta{
							Name:      RandomLowercaseString(),
							Namespace: "argo",
						},
						Spec: pipelineSpec,
					},
					k8sClient, ctx)

				Expect(KfpUploadToFail(testCtx.Pipeline.Name, PipelineId)).To(Succeed())
				Expect(KfpUploadVersionToFail(testCtx.Pipeline.Name, PipelineId)).To(Succeed())

				workflow, err := workflows.ConstructCreationWorkflow(testCtx.Pipeline.Spec, testCtx.Pipeline.ObjectMeta, V1)
				Expect(err).NotTo(HaveOccurred())

				err = k8sClient.Create(ctx, workflow)
				Expect(err).NotTo(HaveOccurred())

				Eventually(testCtx.WorkflowToMatch(CreateOperationLabel, func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
				}), TestTimeout).Should(Succeed())
			})
		})

		When("The creation fails", func() {
			It("Fails the workflow", func() {
				testCtx := NewTestContextWithPipeline(
					&pipelinesv1.Pipeline{
						ObjectMeta: metav1.ObjectMeta{
							Name:      RandomLowercaseString(),
							Namespace: "argo",
						},
						Spec: pipelineSpec,
					},
					k8sClient, ctx)

				Expect(KfpUploadToSucceed(testCtx.Pipeline.Name, PipelineId)).To(Succeed())
				Expect(KfpUploadVersionToFail(PipelineId, V1)).To(Succeed())

				workflow, err := workflows.ConstructCreationWorkflow(testCtx.Pipeline.Spec, testCtx.Pipeline.ObjectMeta, V1)
				Expect(err).NotTo(HaveOccurred())

				err = k8sClient.Create(ctx, workflow)
				Expect(err).NotTo(HaveOccurred())

				Eventually(testCtx.WorkflowToMatch(CreateOperationLabel, func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
				}), TestTimeout).Should(Succeed())
			})
		})
	})

	Describe("Upload workflow", func() {
		When("The upload succeeds", func() {
			It("Succeeds the workflow", func() {
				testCtx := NewTestContextWithPipeline(
					&pipelinesv1.Pipeline{
						ObjectMeta: metav1.ObjectMeta{
							Name:      RandomLowercaseString(),
							Namespace: "argo",
						},
						Spec: pipelineSpec,
					},
					k8sClient, ctx)

				Expect(KfpUploadVersionToReturn(testCtx.Pipeline.Name, PipelineId, V1)).To(Succeed())

				workflow, err := workflows.ConstructUpdateWorkflow(testCtx.Pipeline.Spec, testCtx.Pipeline.ObjectMeta, PipelineId, V1)
				Expect(err).NotTo(HaveOccurred())

				err = k8sClient.Create(ctx, workflow)
				Expect(err).NotTo(HaveOccurred())

				Eventually(testCtx.WorkflowToMatch(UpdateOperationLabel, func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				}), TestTimeout).Should(Succeed())
			})
		})

		When("The upload fails", func() {
			It("Fails the workflow", func() {
				testCtx := NewTestContextWithPipeline(
					&pipelinesv1.Pipeline{
						ObjectMeta: metav1.ObjectMeta{
							Name:      RandomLowercaseString(),
							Namespace: "argo",
						},
						Spec: pipelineSpec,
					},
					k8sClient, ctx)

				Expect(KfpUploadVersionToFail(PipelineId, V1)).To(Succeed())

				workflow, err := workflows.ConstructUpdateWorkflow(testCtx.Pipeline.Spec, testCtx.Pipeline.ObjectMeta, PipelineId, V1)
				Expect(err).NotTo(HaveOccurred())

				err = k8sClient.Create(ctx, workflow)
				Expect(err).NotTo(HaveOccurred())

				Eventually(testCtx.WorkflowToMatch(UpdateOperationLabel, func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
				}), TestTimeout).Should(Succeed())
			})
		})
	})

	Describe("Deletion workflow", func() {
		When("The deletion succeeds", func() {
			It("Succeeds the workflow", func() {
				testCtx := NewTestContextWithPipeline(
					&pipelinesv1.Pipeline{
						ObjectMeta: metav1.ObjectMeta{
							Name:      RandomLowercaseString(),
							Namespace: "argo",
						},
						Spec: pipelineSpec,
						Status: pipelinesv1.PipelineStatus{
							Id: PipelineId,
						},
					},
					k8sClient, ctx)

				Expect(KfpDeleteToReturn(*testCtx.Pipeline, PipelineId)).To(Succeed())

				workflow := workflows.ConstructDeletionWorkflow(testCtx.Pipeline.ObjectMeta, PipelineId)

				err := k8sClient.Create(ctx, workflow)
				Expect(err).NotTo(HaveOccurred())

				Eventually(testCtx.WorkflowToMatch(DeleteOperationLabel, func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				}), TestTimeout).Should(Succeed())
			})
		})

		When("The deletion fails", func() {
			It("Fails the workflow", func() {
				testCtx := NewTestContextWithPipeline(
					&pipelinesv1.Pipeline{
						ObjectMeta: metav1.ObjectMeta{
							Name:      RandomLowercaseString(),
							Namespace: "argo",
						},
						Spec: pipelineSpec,
						Status: pipelinesv1.PipelineStatus{
							Id: PipelineId,
						},
					},
					k8sClient, ctx)

				Expect(KfpDeleteToReturn(*testCtx.Pipeline, PipelineId)).To(Succeed())
				Expect(KfpDeleteToFail(*testCtx.Pipeline, PipelineId)).To(Succeed())

				workflow := workflows.ConstructDeletionWorkflow(testCtx.Pipeline.ObjectMeta, PipelineId)

				err := k8sClient.Create(ctx, workflow)
				Expect(err).NotTo(HaveOccurred())

				Eventually(testCtx.WorkflowToMatch(DeleteOperationLabel, func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
				}), TestTimeout).Should(Succeed())
			})
		})
	})
})
