//go:build integration
// +build integration

package pipelines

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1alpha3"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	"github.com/walkerus/go-wiremock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Context("Pipeline Workflows", Serial, func() {

	workflowFactory := PipelineWorkflowFactory{
		WorkflowFactoryBase: WorkflowFactoryBase{
			Config: configv1.Configuration{
				PipelineStorage: "gs://some-bucket",
				DefaultBeamArgs: []apis.NamedValue{
					{Name: "project", Value: "project"},
				},
				KfpEndpoint:            "http://wiremock:80",
				WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
			},
		},
	}

	var kfpId = RandomString()

	var SucceedUpload = func(pipeline *pipelinesv1.Pipeline) error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload")).
			WithQueryParam("name", wiremock.EqualTo(pipeline.Name)).
			WillReturn(
				fmt.Sprintf(`{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s", "default_version": {"id": "%s"}}`, pipeline.Status.KfpId, pipeline.Name, pipeline.Name),
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
				fmt.Sprintf(`{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s", "resource_references": [{"key": {"id": "%s", "type": "PIPELINE"}, "name": "%s", "relationship": "OWNER"}]}`,
					RandomString(),
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
		constructWorkflow func(*pipelinesv1.Pipeline) (*argo.Workflow, error),
		assertion func(Gomega, *argo.Workflow)) {

		testCtx := NewPipelineTestContext(
			&pipelinesv1.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:      RandomLowercaseString(),
					Namespace: "argo",
				},
				Spec: pipelineSpec,
				Status: apis.Status{
					KfpId: kfpId,
				},
			},
			k8sClient, ctx)

		setUp(testCtx.Pipeline)
		workflow, err := constructWorkflow(testCtx.Pipeline)
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
				g.Expect(getWorkflowOutput(workflow, PipelineWorkflowConstants.PipelineVersionParameterName)).
					To(Equal(pipelineSpec.ComputeVersion()))
			},
		),
		Entry("Creation succeeds but the update fails",
			func(pipeline *pipelinesv1.Pipeline) {
				Expect(SucceedUpload(pipeline)).To(Succeed())
				Expect(FailUploadVersion(pipeline)).To(Succeed())
			},
			workflowFactory.ConstructCreationWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				g.Expect(getWorkflowOutput(workflow, PipelineWorkflowConstants.PipelineIdParameterName)).
					To(Equal(kfpId))
				g.Expect(getWorkflowOutput(workflow, PipelineWorkflowConstants.PipelineVersionParameterName)).
					To(Equal(""))
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

	DescribeTable("Deletion Workflow", AssertWorkflow,
		Entry("Deletion succeeds",
			func(pipeline *pipelinesv1.Pipeline) {
				Expect(SucceedDeletion(pipeline)).To(Succeed())
			},
			func(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
				return workflowFactory.ConstructDeletionWorkflow(pipeline)
			},
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
			},
		),
		Entry("Deletion fails",
			func(pipeline *pipelinesv1.Pipeline) {
				Expect(FailDeletion(pipeline)).To(Succeed())
			},
			func(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
				return workflowFactory.ConstructDeletionWorkflow(pipeline)
			},
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
			},
		),
	)
})
