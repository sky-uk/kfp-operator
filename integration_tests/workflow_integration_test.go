package integration

import (
	"context"
	"testing"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	. "github.com/sky-uk/kfp-operator/controllers/pipelines/test_utils"
	pipelineWorkflows "github.com/sky-uk/kfp-operator/controllers/pipelines/workflows"
	"github.com/sky-uk/kfp-operator/external"
	"github.com/walkerus/go-wiremock"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestBooks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Argo Integration Suite")
}

const (
	TestTimeout = 120
)

var (
	restCfg = rest.Config{
		Host:    "http://localhost:8080",
		APIPath: "/api",
	}

	wiremockClient *wiremock.Client
	k8sClient      client.Client
	ctx            context.Context
	workflows      pipelineWorkflows.Workflows
)

var _ = BeforeSuite(func() {
	wiremockClient = wiremock.NewClient("http://localhost:8081")

	Expect(external.InitSchemes(scheme.Scheme)).To(Succeed())
	var err error
	k8sClient, err = client.New(&restCfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	ctx = context.Background()

	workflows = pipelineWorkflows.Workflows{
		Config: pipelineWorkflows.Configuration{
			KfpEndpoint:     "http://kfp-wiremock:80",
			KfpToolsImage:   "kfp-tools",
			CompilerImage:   "compiler",
			ImagePullPolicy: "Never", // Needed for minikube to use local images
			PipelineStorage: "gs://some-bucket",
			DataflowProject: "project",
		},
	}
})

var _ = BeforeEach(func() {
	Expect(wiremockClient.Reset()).To(Succeed())

	Expect(wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/healthz")).
		WillReturn(
			`{"status": "ok"}`,
			map[string]string{"Content-Type": "application/json"},
			200,
		))).To(Succeed())
})

func KfpUploadToSucceed(pipelineName string, pipelineId string) error {
	return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload")).
		WithQueryParam("name", wiremock.EqualTo(pipelineName)).
		WillReturn(
			`{"id": "`+pipelineId+`", "created_at": "2021-09-10T15:46:08Z", "name": "`+pipelineName+`"}`,
			map[string]string{"Content-Type": "application/json"},
			200,
		))
}

func KfpUploadToFail(pipelineName string, pipelineId string) error {
	return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload")).
		WithQueryParam("name", wiremock.EqualTo(pipelineName)).
		WillReturn(
			`{"status": "failed"}`,
			map[string]string{"Content-Type": "application/json"},
			404,
		))
}

func KfpUploadVersionToReturn(pipelineName string, pipelineId string, pipelineVersion string) error {
	return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload_version")).
		WithQueryParam("name", wiremock.EqualTo(pipelineVersion)).
		WithQueryParam("pipelineid", wiremock.EqualTo(pipelineId)).
		WillReturn(
			`{"id": "`+pipelineVersion+`", "created_at": "2021-09-10T15:46:08Z", "name": "pipeline", "resource_references": [{"key": {"id": "`+pipelineId+`", "apiResourceType": "PIPELINE"}, "name": "`+pipelineName+`", "relationship": "OWNER"}]}`,
			map[string]string{"Content-Type": "application/json"},
			200,
		))
}

func KfpUploadVersionToFail(pipelineId string, pipelineVersion string) error {
	return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload_version")).
		WithQueryParam("name", wiremock.EqualTo(pipelineVersion)).
		WithQueryParam("pipelineid", wiremock.EqualTo(pipelineId)).
		WillReturn(
			`{"status": "failed"`,
			map[string]string{"Content-Type": "application/json"},
			400,
		))
}

func KfpDeleteToReturn(pipeline pipelinesv1.Pipeline, pipelineId string) error {
	return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/"+pipelineId)).
		WillReturn(
			`{"satus": "deleted"}`,
			map[string]string{"Content-Type": "application/json"},
			200,
		))
}

func KfpDeleteToFail(pipeline pipelinesv1.Pipeline, pipelineId string) error {
	return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/"+pipelineId)).
		WillReturn(
			`{"satus": "failed"}`,
			map[string]string{"Content-Type": "application/json"},
			400,
		))
}

var _ = Describe("Creation workflow", func() {
	When("The creation and update succeeds", func() {
		It("Succeeds the workflow with a Pipeline Id", func() {

			testCtx := NewTestContext(k8sClient, ctx)
			testCtx.Pipeline.ObjectMeta.Namespace = "argo"

			Expect(KfpUploadToSucceed(testCtx.Pipeline.Name, PipelineId)).To(Succeed())
			Expect(KfpUploadVersionToReturn(testCtx.Pipeline.Name, PipelineId, V1)).To(Succeed())

			workflow, err := workflows.ConstructCreationWorkflow(testCtx.Pipeline.Spec, testCtx.Pipeline.ObjectMeta, V1)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Create(ctx, workflow)
			Expect(err).NotTo(HaveOccurred())

			Eventually(testCtx.WorkflowToMatch(pipelineWorkflows.Create, func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				g.Expect(pipelineWorkflows.GetWorkflowOutput(workflow, pipelineWorkflows.PipelineIdParameterName)).
					To(Equal(PipelineId))
			}), TestTimeout).Should(Succeed())
		})
	})

	/* Not currently supported because argo workflow does not fail with output
	When("The creation succeeds but the update fails", func() {
		It("Fails the workflow with a Pipeline Id", func() {
			pipelineId := "12345"
			pipelineVersion := "abcdef"
			testCtx := testutils.NewTestContext(k8sClient, ctx)
			testCtx.Pipeline.Status.Version = pipelineVersion

			Expect(KfpUploadToReturn(*testCtx.Pipeline, pipelineId)).To(Succeed())
			Expect(KfpUploadVersionToFail(*testCtx.Pipeline, pipelineId)).To(Succeed())

			workflow, err := workflows.ConstructCreationWorkflow(testCtx.Pipeline)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Create(ctx, workflow)
			Expect(err).NotTo(HaveOccurred())

			Eventually(testCtx.WorkflowToMatch(pipelineWorkflows.Create, func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
				g.Expect(pipelineWorkflows.GetWorkflowOutput(workflow, pipelineWorkflows.PipelineIdParameterName)).To(Equal(pipelineId))
			}), TestTimeout).Should(Succeed())
		})
	})
	*/

	When("The creation fails", func() {
		It("Fails the workflow", func() {
			testCtx := NewTestContext(k8sClient, ctx)
			testCtx.Pipeline.ObjectMeta.Namespace = "argo"

			Expect(KfpUploadToSucceed(testCtx.Pipeline.Name, PipelineId)).To(Succeed())
			Expect(KfpUploadVersionToFail(PipelineId, V1)).To(Succeed())

			workflow, err := workflows.ConstructCreationWorkflow(testCtx.Pipeline.Spec, testCtx.Pipeline.ObjectMeta, V1)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Create(ctx, workflow)
			Expect(err).NotTo(HaveOccurred())

			Eventually(testCtx.WorkflowToMatch(pipelineWorkflows.Create, func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
			}), TestTimeout).Should(Succeed())
		})
	})
})

var _ = Describe("Upload workflow", func() {
	When("The upload succeeds", func() {
		It("Succeeds the workflow", func() {
			testCtx := NewTestContext(k8sClient, ctx)
			testCtx.Pipeline.ObjectMeta.Namespace = "argo"

			Expect(KfpUploadVersionToReturn(testCtx.Pipeline.Name, PipelineId, V1)).To(Succeed())

			workflow, err := workflows.ConstructUpdateWorkflow(testCtx.Pipeline.Spec, testCtx.Pipeline.ObjectMeta, PipelineId, V1)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Create(ctx, workflow)
			Expect(err).NotTo(HaveOccurred())

			Eventually(testCtx.WorkflowToMatch(pipelineWorkflows.Update, func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
			}), TestTimeout).Should(Succeed())
		})
	})

	When("The upload fails", func() {
		It("Fails the workflow", func() {
			testCtx := NewTestContext(k8sClient, ctx)
			testCtx.Pipeline.ObjectMeta.Namespace = "argo"

			Expect(KfpUploadVersionToFail(PipelineId, V1)).To(Succeed())

			workflow, err := workflows.ConstructUpdateWorkflow(testCtx.Pipeline.Spec, testCtx.Pipeline.ObjectMeta, PipelineId, V1)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Create(ctx, workflow)
			Expect(err).NotTo(HaveOccurred())

			Eventually(testCtx.WorkflowToMatch(pipelineWorkflows.Update, func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
			}), TestTimeout).Should(Succeed())
		})
	})
})

var _ = Describe("Deletion workflow", func() {
	When("The deletion succeeds", func() {
		It("Succeeds the workflow", func() {
			testCtx := NewTestContext(k8sClient, ctx)
			testCtx.Pipeline.Status.Id = PipelineId
			testCtx.Pipeline.ObjectMeta.Namespace = "argo"

			Expect(KfpDeleteToReturn(*testCtx.Pipeline, PipelineId)).To(Succeed())

			workflow := workflows.ConstructDeletionWorkflow(testCtx.Pipeline.ObjectMeta, PipelineId)

			err := k8sClient.Create(ctx, workflow)
			Expect(err).NotTo(HaveOccurred())

			Eventually(testCtx.WorkflowToMatch(pipelineWorkflows.Delete, func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
			}), TestTimeout).Should(Succeed())
		})
	})

	When("The deletion fails", func() {
		It("Fails the workflow", func() {
			testCtx := NewTestContext(k8sClient, ctx)
			testCtx.Pipeline.Status.Id = PipelineId
			testCtx.Pipeline.ObjectMeta.Namespace = "argo"

			Expect(KfpDeleteToReturn(*testCtx.Pipeline, PipelineId)).To(Succeed())
			Expect(KfpDeleteToFail(*testCtx.Pipeline, PipelineId)).To(Succeed())

			workflow := workflows.ConstructDeletionWorkflow(testCtx.Pipeline.ObjectMeta, PipelineId)

			err := k8sClient.Create(ctx, workflow)
			Expect(err).NotTo(HaveOccurred())

			Eventually(testCtx.WorkflowToMatch(pipelineWorkflows.Delete, func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
			}), TestTimeout).Should(Succeed())
		})
	})
})
