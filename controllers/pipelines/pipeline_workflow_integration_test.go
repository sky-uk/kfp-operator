//go:build integration
// +build integration

package pipelines

import (
	"context"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"github.com/walkerus/go-wiremock"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Context("Pipeline Workflows", func() {

	workflowFactory := PipelineWorkflowFactory{
		WorkflowFactory: WorkflowFactory{
			Config: configv1.Configuration{
				KfpEndpoint: "http://wiremock:80",
				Argo: configv1.ArgoConfiguration{
					KfpSdkImage:   "kfp-operator-argo-kfp-sdk",
					CompilerImage: "kfp-operator-argo-kfp-compiler",
					ContainerDefaults: apiv1.Container{
						ImagePullPolicy: "Never", // Needed for minikube to use local images
					},
				},
				PipelineStorage: "gs://some-bucket",
				DefaultBeamArgs: map[string]string{
					"project": "project",
				},
			},
		},
	}

	var kfpId = RandomString()

	var SucceedUpload = func(pipeline *pipelinesv1.Pipeline) error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload")).
			WithQueryParam("name", wiremock.EqualTo(pipeline.Name)).
			WillReturn(
				fmt.Sprintf(`{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}`, pipeline.Status.KfpId, pipeline.Spec.ComputeVersion()),
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var FailUpload = func(pipeline *pipelinesv1.Pipeline) error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload")).
			WithQueryParam("name", wiremock.EqualTo(pipeline.Name)).
			WillReturn(
				`{"status": "failed"}`,
				map[string]string{"Content-Type": "application/json"},
				404,
			))
	}

	var SucceedUploadVersion = func(pipeline *pipelinesv1.Pipeline) error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload_version")).
			WithQueryParam("name", wiremock.EqualTo(pipeline.Spec.ComputeVersion())).
			WithQueryParam("pipelineid", wiremock.EqualTo(pipeline.Status.KfpId)).
			WillReturn(
				fmt.Sprintf(`{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "pipeline", "resource_references": [{"key": {"id": "%s", "apiResourceType": "PIPELINE"}, "name": "%s", "relationship": "OWNER"}]}`,
					pipeline.Spec.ComputeVersion(),
					pipeline.Status.KfpId,
					pipeline.Name),
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var FailUploadVersion = func(pipeline *pipelinesv1.Pipeline) error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload_version")).
			WithQueryParam("name", wiremock.EqualTo(pipeline.Spec.ComputeVersion())).
			WithQueryParam("pipelineid", wiremock.EqualTo(pipeline.Status.KfpId)).
			WillReturn(
				`{"status": "failed"`,
				map[string]string{"Content-Type": "application/json"},
				400,
			))
	}

	var SucceedDeletion = func(pipeline *pipelinesv1.Pipeline) error {
		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/"+pipeline.Status.KfpId)).
			WillReturn(
				`{"status": "deleted"}`,
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var FailDeletion = func(pipeline *pipelinesv1.Pipeline) error {
		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/"+pipeline.Status.KfpId)).
			WillReturn(
				`{"status": "failed"}`,
				map[string]string{"Content-Type": "application/json"},
				400,
			))
	}

	var AssertWorkflow = func(
		setUp func(pipeline *pipelinesv1.Pipeline),
		constructWorkflow func(context.Context, *pipelinesv1.Pipeline) (*argo.Workflow, error),
		assertion func(Gomega, *argo.Workflow)) {

		testCtx := NewPipelineTestContext(
			&pipelinesv1.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:      RandomLowercaseString(),
					Namespace: "argo",
				},
				Spec: pipelineSpec,
				Status: pipelinesv1.Status{
					KfpId: kfpId,
				},
			},
			k8sClient, ctx)

		setUp(testCtx.Pipeline)
		workflow, err := constructWorkflow(testCtx.ctx, testCtx.Pipeline)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient.Create(ctx, workflow)).To(Succeed())

		Eventually(testCtx.WorkflowByNameToMatch(types.NamespacedName{Name: workflow.Name, Namespace: workflow.Namespace},
			assertion), TestTimeout).Should(Succeed())
	}

	DescribeTable("Creation Workflow", AssertWorkflow,
		Entry("Creation succeeds",
			func(pipeline *pipelinesv1.Pipeline) {
				Expect(SucceedUpload(pipeline)).To(Succeed())
				Expect(SucceedUploadVersion(pipeline)).To(Succeed())
			},
			workflowFactory.ConstructCreationWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				g.Expect(getWorkflowOutput(workflow, PipelineWorkflowConstants.PipelineIdParameterName)).
					To(Equal(kfpId))
			},
		),
		Entry("Creation succeeds but the update fails",
			func(pipeline *pipelinesv1.Pipeline) {
				Expect(SucceedUpload(pipeline)).To(Succeed())
				Expect(FailUploadVersion(pipeline)).To(Succeed())
			},
			workflowFactory.ConstructCreationWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
			},
		),
		Entry("Creation fails",
			func(pipeline *pipelinesv1.Pipeline) {
				Expect(FailUpload(pipeline)).To(Succeed())
				Expect(FailUploadVersion(pipeline)).To(Succeed())
			},
			workflowFactory.ConstructCreationWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
			},
		),
	)

	DescribeTable("Update Workflow", AssertWorkflow,
		Entry("Upload succeeds",
			func(pipeline *pipelinesv1.Pipeline) {
				Expect(SucceedUploadVersion(pipeline)).To(Succeed())
			},
			workflowFactory.ConstructUpdateWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
			},
		),
		Entry("Upload fails",
			func(pipeline *pipelinesv1.Pipeline) {
				Expect(FailUploadVersion(pipeline)).To(Succeed())
			},
			workflowFactory.ConstructUpdateWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
			},
		),
	)

	DescribeTable("Update Workflow", AssertWorkflow,
		Entry("Deletion succeeds",
			func(pipeline *pipelinesv1.Pipeline) {
				Expect(SucceedDeletion(pipeline)).To(Succeed())
			},
			func(ctx context.Context, pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
				return workflowFactory.ConstructDeletionWorkflow(context.Background(), pipeline), nil
			},
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
			},
		),
		Entry("Deletion fails",
			func(pipeline *pipelinesv1.Pipeline) {
				Expect(FailDeletion(pipeline)).To(Succeed())
			},
			func(ctx context.Context, pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
				return workflowFactory.ConstructDeletionWorkflow(context.Background(), pipeline), nil
			},
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
			},
		),
	)
})
