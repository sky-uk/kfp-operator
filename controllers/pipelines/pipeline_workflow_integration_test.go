//go:build integration
// +build integration

package pipelines

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha3"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	"github.com/walkerus/go-wiremock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type PipelineWorkflowIntegrationSuite interface {
	SucceedUpload(*pipelinesv1.Pipeline) error
	FailUpload(*pipelinesv1.Pipeline) error
	SucceedUploadVersion(*pipelinesv1.Pipeline) error
	FailUploadVersion(*pipelinesv1.Pipeline) error
	SucceedDeletion(*pipelinesv1.Pipeline) error
	FailDeletion(*pipelinesv1.Pipeline) error

	ProviderConfig() string
}

type VertexAIPipelineWorkflowIntegrationSuite struct {
}

func (vaipwis VertexAIPipelineWorkflowIntegrationSuite) ProviderConfig() string {
	return "endpoint: http://wiremock:80\nimage: kfp-operator-vai-provider\npipelineBucket: pipelineBucket"
}

func (vaipwis VertexAIPipelineWorkflowIntegrationSuite) SucceedUpload(_ *pipelinesv1.Pipeline) error {
	return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/upload/storage/v1/b/pipelineBucket/o")).
		WithQueryParam("name", wiremock.Matching(".+/")).
		WillReturn("{}",
			map[string]string{"Content-Type": "application/json"},
			200))
}

func (vaipwis VertexAIPipelineWorkflowIntegrationSuite) FailUpload(_ *pipelinesv1.Pipeline) error {
	return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/upload/storage/v1/b/pipelineBucket/o")).
		WithQueryParam("name", wiremock.Matching(".+/")).
		WillReturn("",
			map[string]string{},
			404))
}

func (vaipwis VertexAIPipelineWorkflowIntegrationSuite) SucceedUploadVersion(_ *pipelinesv1.Pipeline) error {
	return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/upload/storage/v1/b/pipelineBucket/o")).
		WithQueryParam("name", wiremock.Matching(".+/.+")).
		WillReturn("{}",
			map[string]string{"Content-Type": "application/json"},
			200))
}

func (vaipwis VertexAIPipelineWorkflowIntegrationSuite) FailUploadVersion(_ *pipelinesv1.Pipeline) error {
	return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/upload/storage/v1/b/pipelineBucket/o")).
		WithQueryParam("name", wiremock.Matching(".+/.+")).
		WillReturn("",
			map[string]string{},
			404))
}

func (vaipwis VertexAIPipelineWorkflowIntegrationSuite) SucceedDeletion(pipeline *pipelinesv1.Pipeline) error {
	err := wiremockClient.StubFor(wiremock.Get(wiremock.URLPathMatching("/b/pipelineBucket/o")).
		WillReturn("{}",
			map[string]string{"Content-Type": "application/json"},
			200))
	if err != nil {
		return err
	}

	return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathMatching("/b/pipelineBucket/o/.*")).
		WillReturn("{}",
			map[string]string{"Content-Type": "application/json"},
			200))
}

func (vaipwis VertexAIPipelineWorkflowIntegrationSuite) FailDeletion(pipeline *pipelinesv1.Pipeline) error {
	return wiremockClient.StubFor(wiremock.Get(wiremock.URLPathMatching("/b/pipelineBucket/o")).
		WillReturn("{}",
			map[string]string{"Content-Type": "application/json"},
			404))
}

type KfpPipelineWorkflowIntegrationSuite struct {
}

func (kfppwis KfpPipelineWorkflowIntegrationSuite) SucceedUpload(pipeline *pipelinesv1.Pipeline) error {
	return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload")).
		WithQueryParam("name", wiremock.EqualTo(pipeline.Name)).
		WillReturn(
			fmt.Sprintf(`{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s", "default_version": {"id": "%s"}}`, pipeline.Status.KfpId, pipeline.Name, pipeline.Name),
			map[string]string{"Content-Type": "application/json"},
			200,
		))
}

func (kfppwis KfpPipelineWorkflowIntegrationSuite) FailUpload(pipeline *pipelinesv1.Pipeline) error {
	return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload")).
		WithQueryParam("name", wiremock.EqualTo(pipeline.Name)).
		WillReturn(
			`{"status": "failed"}`,
			map[string]string{"Content-Type": "application/json"},
			404,
		))
}

func (kfppwis KfpPipelineWorkflowIntegrationSuite) SucceedUploadVersion(pipeline *pipelinesv1.Pipeline) error {
	return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload_version")).
		WithQueryParam("name", wiremock.EqualTo(pipeline.Spec.ComputeVersion())).
		WithQueryParam("pipelineid", wiremock.EqualTo(pipeline.Status.KfpId)).
		WillReturn(
			fmt.Sprintf(`{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s", "resource_references": [{"key": {"id": "%s", "type": "PIPELINE"}, "name": "%s", "relationship": "OWNER"}]}`,
				apis.RandomString(),
				pipeline.Spec.ComputeVersion(),
				pipeline.Status.KfpId,
				pipeline.Name),
			map[string]string{"Content-Type": "application/json"},
			200,
		))
}

func (kfppwis KfpPipelineWorkflowIntegrationSuite) SucceedDeletion(pipeline *pipelinesv1.Pipeline) error {
	return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/"+pipeline.Status.KfpId)).
		WillReturn(
			`{"status": "deleted"}`,
			map[string]string{"Content-Type": "application/json"},
			200,
		))
}

func (kfppwis KfpPipelineWorkflowIntegrationSuite) FailDeletion(pipeline *pipelinesv1.Pipeline) error {
	return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/"+pipeline.Status.KfpId)).
		WillReturn(
			`{"status": "failed"}`,
			map[string]string{"Content-Type": "application/json"},
			400,
		))
}

func (kfppwis KfpPipelineWorkflowIntegrationSuite) FailUploadVersion(pipeline *pipelinesv1.Pipeline) error {
	return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines/upload_version")).
		WithQueryParam("name", wiremock.EqualTo(pipeline.Spec.ComputeVersion())).
		WithQueryParam("pipelineid", wiremock.EqualTo(pipeline.Status.KfpId)).
		WillReturn(
			`{"status": "failed"`,
			map[string]string{"Content-Type": "application/json"},
			400,
		))
}

func (kfppwis KfpPipelineWorkflowIntegrationSuite) ProviderConfig() string {
	return "endpoint: http://wiremock:80\nimage: kfp-operator-kfp-provider"
}

var kfpId = apis.RandomString()

func AssertWorkflow(
	setUp func(pipeline *pipelinesv1.Pipeline),
	constructWorkflow func(*pipelinesv1.Pipeline) (*argo.Workflow, error),
	assertion func(Gomega, *argo.Workflow)) {

	testCtx := NewPipelineTestContext(
		&pipelinesv1.Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name:      apis.RandomLowercaseString(),
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

func RunSuite(suite PipelineWorkflowIntegrationSuite, suitName string) {
	Context(suitName, func() {
		workflowFactory := PipelineWorkflowFactory{
			WorkflowFactoryBase: WorkflowFactoryBase{
				Config: config.Configuration{
					PipelineStorage: "gs://some-bucket",
					DefaultBeamArgs: []apis.NamedValue{
						{Name: "project", Value: "project"},
					},
					WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
				},
				ProviderConfig: suite.ProviderConfig(),
			},
		}

		DescribeTable("Creation Workflow", AssertWorkflow,
			Entry("Creation succeeds",
				func(pipeline *pipelinesv1.Pipeline) {
					Expect(suite.SucceedUpload(pipeline)).To(Succeed())
					Expect(suite.SucceedUploadVersion(pipeline)).To(Succeed())
				},
				workflowFactory.ConstructCreationWorkflow,
				func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
					output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(output.Id).NotTo(BeEmpty())
					g.Expect(output.ProviderError).To(BeEmpty())
				},
			),
			Entry("Creation succeeds but the update fails",
				func(pipeline *pipelinesv1.Pipeline) {
					Expect(suite.SucceedUpload(pipeline)).To(Succeed())
					Expect(suite.FailUploadVersion(pipeline)).To(Succeed())
				},
				workflowFactory.ConstructCreationWorkflow,
				func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
					output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(output.Id).NotTo(BeEmpty())
					g.Expect(output.ProviderError).NotTo(BeEmpty())
				},
			),
			Entry("Creation fails",
				func(pipeline *pipelinesv1.Pipeline) {
					Expect(suite.FailUpload(pipeline)).To(Succeed())
					Expect(suite.FailUploadVersion(pipeline)).To(Succeed())
				},
				workflowFactory.ConstructCreationWorkflow,
				func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
					output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(output.Id).To(BeEmpty())
					g.Expect(output.ProviderError).NotTo(BeEmpty())
				},
			),
		)

		DescribeTable("Update Workflow", AssertWorkflow,
			Entry("Update succeeds",
				func(pipeline *pipelinesv1.Pipeline) {
					Expect(suite.SucceedUploadVersion(pipeline)).To(Succeed())
				},
				workflowFactory.ConstructUpdateWorkflow,
				func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
					output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(output.Id).NotTo(BeEmpty())
					g.Expect(output.ProviderError).To(BeEmpty())
				},
			),
			Entry("Update fails",
				func(pipeline *pipelinesv1.Pipeline) {
					Expect(suite.FailUploadVersion(pipeline)).To(Succeed())
				},
				workflowFactory.ConstructUpdateWorkflow,
				func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
					output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(output.Id).To(Equal(kfpId))
					g.Expect(output.ProviderError).NotTo(BeEmpty())
				},
			),
		)

		DescribeTable("Deletion Workflow", AssertWorkflow,
			Entry("Deletion succeeds",
				func(pipeline *pipelinesv1.Pipeline) {
					Expect(suite.SucceedDeletion(pipeline)).To(Succeed())
				},
				func(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
					return workflowFactory.ConstructDeletionWorkflow(pipeline)
				},
				func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
					output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(output.Id).To(BeEmpty())
					g.Expect(output.ProviderError).To(BeEmpty())
				},
			),
			Entry("Deletion fails",
				func(pipeline *pipelinesv1.Pipeline) {
					Expect(suite.FailDeletion(pipeline)).To(Succeed())
				},
				func(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
					return workflowFactory.ConstructDeletionWorkflow(pipeline)
				},
				func(g Gomega, workflow *argo.Workflow) {
					g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
					output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(output.Id).To(Equal(kfpId))
					g.Expect(output.ProviderError).NotTo(BeEmpty())
				},
			),
		)
	})
}

var _ = Context("Pipeline Workflows", Serial, func() {
	RunSuite(VertexAIPipelineWorkflowIntegrationSuite{}, "Vertex AI")
	RunSuite(KfpPipelineWorkflowIntegrationSuite{}, "Kubeflow Pipelines")
})
