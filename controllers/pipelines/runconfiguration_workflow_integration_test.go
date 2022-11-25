//go:build integration
// +build integration

package pipelines

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"github.com/walkerus/go-wiremock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type RunConfigurationWorkflowIntegrationSuite interface {
	SucceedCreation(runConfiguration *pipelinesv1.RunConfiguration) error
	FailCreation(runConfiguration *pipelinesv1.RunConfiguration) error

	SucceedUpdate(runConfiguration *pipelinesv1.RunConfiguration) error
	SucceedUpdateNotFound(runConfiguration *pipelinesv1.RunConfiguration) error
	FailRecreate(runConfiguration *pipelinesv1.RunConfiguration) error // TODO Skip for VAI
	FailUpdate(runConfiguration *pipelinesv1.RunConfiguration) error

	SucceedDeletion(runConfiguration *pipelinesv1.RunConfiguration) error
	FailDeletion(runConfiguration *pipelinesv1.RunConfiguration) error
	FailDeletionNotFound(runConfiguration *pipelinesv1.RunConfiguration) error

	ProviderName() string
}

type KfpRunConfigurationWorkflowIntegrationSuite struct {
}

func (kfprcwis KfpRunConfigurationWorkflowIntegrationSuite) SucceedCreation(runConfiguration *pipelinesv1.RunConfiguration) error {
	return kfprcwis.succeedCreation(runConfiguration, runConfigurationProviderId)
}

func (kfprcwis KfpRunConfigurationWorkflowIntegrationSuite) succeedCreation(runConfiguration *pipelinesv1.RunConfiguration, createdProviderId string) error {
	if err := kfprcwis.stubGetExperiment(defaultExperiment); err != nil {
		return err
	}

	if err := kfprcwis.stubGetPipeline(runConfiguration.Spec.Pipeline.Name); err != nil {
		return err
	}
	if err := kfprcwis.stubGetPipelineVersions(); err != nil {
		return err
	}

	return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/jobs")).
		WillReturn(
			fmt.Sprintf(`{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}`,
				createdProviderId, runConfiguration.Name),
			map[string]string{"Content-Type": "application/json"},
			200,
		))
}

func (_ KfpRunConfigurationWorkflowIntegrationSuite) FailCreation(_ *pipelinesv1.RunConfiguration) error {
	return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/jobs")).
		WillReturn(
			`{"status": "failed"}`,
			map[string]string{"Content-Type": "application/json"},
			404,
		))
}

func (kfprcwis KfpRunConfigurationWorkflowIntegrationSuite) SucceedUpdate(runConfiguration *pipelinesv1.RunConfiguration) error {
	if err := kfprcwis.SucceedDeletion(runConfiguration); err != nil {
		return err
	}
	if err := kfprcwis.stubGetPipelineVersions(); err != nil {
		return err
	}
	return kfprcwis.succeedCreation(runConfiguration, newRunConfigurationProviderId)
}

func (kfprcwis KfpRunConfigurationWorkflowIntegrationSuite) SucceedUpdateNotFound(runConfiguration *pipelinesv1.RunConfiguration) error {
	if err := kfprcwis.stubGetPipelineVersions(); err != nil {
		return err
	}
	if err := kfprcwis.FailDeletionNotFound(runConfiguration); err != nil {
		return err
	}
	return kfprcwis.succeedCreation(runConfiguration, newRunConfigurationProviderId)
}

func (kfprcwis KfpRunConfigurationWorkflowIntegrationSuite) FailRecreate(runConfiguration *pipelinesv1.RunConfiguration) error {
	if err := kfprcwis.SucceedDeletion(runConfiguration); err != nil {
		return err
	}
	return kfprcwis.FailCreation(runConfiguration)
}

func (kfprcwis KfpRunConfigurationWorkflowIntegrationSuite) FailUpdate(runConfiguration *pipelinesv1.RunConfiguration) error {
	return kfprcwis.FailDeletion(runConfiguration)
}

func (_ KfpRunConfigurationWorkflowIntegrationSuite) SucceedDeletion(_ *pipelinesv1.RunConfiguration) error {
	return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/jobs/"+runConfigurationProviderId)).
		WillReturn(
			`{"status": "deleted"}`,
			map[string]string{"Content-Type": "application/json"},
			200,
		))
}

func (kfprcwis KfpRunConfigurationWorkflowIntegrationSuite) FailDeletion(runConfiguration *pipelinesv1.RunConfiguration) error {
	return kfprcwis.failDeletionWithCode(7)
}

func (kfprcwis KfpRunConfigurationWorkflowIntegrationSuite) FailDeletionNotFound(runConfiguration *pipelinesv1.RunConfiguration) error {
	return kfprcwis.failDeletionWithCode(5)
}

func (_ KfpRunConfigurationWorkflowIntegrationSuite) failDeletionWithCode(code int64) error {
	return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/jobs/"+runConfigurationProviderId)).
		WillReturn(
			fmt.Sprintf(`HTTP response body: {"status": "failed", "code": %d}`, code),
			map[string]string{"Content-Type": "application/json"},
			404,
		))
}

func (_ KfpRunConfigurationWorkflowIntegrationSuite) ProviderName() string {
	return "kfp"
}

func (kfprcwis KfpRunConfigurationWorkflowIntegrationSuite) stubGetPipeline(pipelineName string) error {
	return wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines")).
		WithQueryParam("filter", wiremock.EqualToJson(
			fmt.Sprintf(`{"predicates": [{"op": 1, "key": "name", "stringValue": "%s"}]}`, pipelineName))).
		WillReturn(
			fmt.Sprintf(`{"pipelines": [{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}]}`,
				pipelineProviderId, pipelineName),
			map[string]string{"Content-Type": "application/json"},
			200,
		))
}

func (_ KfpRunConfigurationWorkflowIntegrationSuite) stubGetPipelineVersions() error {
	return wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/pipeline_versions")).
		WithQueryParam("resource_key.id", wiremock.EqualTo(pipelineProviderId)).
		WithQueryParam("filter", wiremock.EqualToJson(
			fmt.Sprintf(`{"predicates": [{"op": "EQUALS", "key": "name", "string_value": "%s"}]}`, versionName))).
		WillReturn(
			fmt.Sprintf(`{"versions": [{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}]}`,
				versionProviderId, versionName),
			map[string]string{"Content-Type": "application/json"},
			200,
		))
}

func (_ KfpRunConfigurationWorkflowIntegrationSuite) stubGetExperiment(experimentName string) error {
	return wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/experiments")).
		WithQueryParam("filter", wiremock.EqualToJson(
			fmt.Sprintf(`{"predicates": [{"op": 1, "key": "name", "stringValue": "%s"}]}`, experimentName))).
		WillReturn(
			fmt.Sprintf(`{"experiments": [{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}]}`,
				experimentProviderId, experimentName),
			map[string]string{"Content-Type": "application/json"},
			200,
		))
}

var pipelineProviderId = apis.RandomString()
var versionProviderId = apis.RandomString()
var versionName = apis.RandomString()
var runConfigurationProviderId = apis.RandomString()    // TODO runConfigurationProviderIdBefore
var newRunConfigurationProviderId = apis.RandomString() // TODO runConfigurationProviderIdAfter
var experimentProviderId = apis.RandomString()
var defaultExperiment = apis.RandomString()

var _ = Context("RunConfiguration Workflows", Serial, func() {

	var AssertWorkflow = func(
		setUp func(runConfiguration *pipelinesv1.RunConfiguration),
		constructWorkflow func(*pipelinesv1.RunConfiguration) (*argo.Workflow, error),
		assertion func(Gomega, *argo.Workflow)) {

		testCtx := NewRunConfigurationTestContext(
			&pipelinesv1.RunConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apis.RandomLowercaseString(),
					Namespace: "argo",
				},
				Spec: pipelinesv1.RunConfigurationSpec{
					Pipeline: pipelinesv1.PipelineIdentifier{Name: apis.RandomString(), Version: versionName},
					Schedule: "* * * * * *",
				},
				Status: pipelinesv1.RunConfigurationStatus{
					Status: pipelinesv1.Status{
						ProviderId: runConfigurationProviderId,
					},
					ObservedPipelineVersion: versionName,
				},
			},
			k8sClient, ctx)

		setUp(testCtx.RunConfiguration)
		workflow, err := constructWorkflow(testCtx.RunConfiguration)

		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient.Create(ctx, workflow)).To(Succeed())

		Eventually(testCtx.WorkflowByNameToMatch(types.NamespacedName{Name: workflow.Name, Namespace: workflow.Namespace},
			assertion), TestTimeout).Should(Succeed())
	}

	var RunSuite = func(suite RunConfigurationWorkflowIntegrationSuite, suitName string) {
		Context(suitName, func() {
			workflowFactory := RunConfigurationWorkflowFactory{
				WorkflowFactoryBase: WorkflowFactoryBase{
					Config: config.Configuration{
						DefaultProviderName:    suite.ProviderName(),
						WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
						DefaultExperiment:      defaultExperiment,
						WorkflowNamespace:      "argo",
					},
				},
			}

			DescribeTable("Creation Workflow", AssertWorkflow,
				Entry("Creation succeeds",
					func(runConfiguration *pipelinesv1.RunConfiguration) {
						Expect(suite.SucceedCreation(runConfiguration)).To(Succeed())
					},
					workflowFactory.ConstructCreationWorkflow,
					func(g Gomega, workflow *argo.Workflow) {
						g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
						output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
						g.Expect(err).NotTo(HaveOccurred())
						g.Expect(output.Id).To(Equal(runConfigurationProviderId))
						g.Expect(output.ProviderError).To(BeEmpty())
					},
				),
				Entry("Creation fails",
					func(runConfiguration *pipelinesv1.RunConfiguration) {
						Expect(suite.FailCreation(runConfiguration)).To(Succeed())
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
				Entry("Deletion and creation succeed", func(runConfiguration *pipelinesv1.RunConfiguration) {
					Expect(suite.SucceedUpdate(runConfiguration)).To(Succeed())
				},
					workflowFactory.ConstructUpdateWorkflow,
					func(g Gomega, workflow *argo.Workflow) {
						g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
						output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
						g.Expect(err).NotTo(HaveOccurred())
						g.Expect(output.Id).To(Equal(newRunConfigurationProviderId))
						g.Expect(output.ProviderError).To(BeEmpty())
					}),
				Entry("Deletion succeeds and creation fails", func(runConfiguration *pipelinesv1.RunConfiguration) {
					Expect(suite.FailRecreate(runConfiguration)).To(Succeed())
				},
					workflowFactory.ConstructUpdateWorkflow,
					func(g Gomega, workflow *argo.Workflow) {
						g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
						output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
						g.Expect(err).NotTo(HaveOccurred())
						g.Expect(output.Id).To(BeEmpty())
						g.Expect(output.ProviderError).NotTo(BeEmpty())
					}),
				Entry("Deletion fails", func(runConfiguration *pipelinesv1.RunConfiguration) {
					Expect(suite.FailUpdate(runConfiguration)).To(Succeed())
				},
					workflowFactory.ConstructUpdateWorkflow,
					func(g Gomega, workflow *argo.Workflow) {
						g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
						output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
						g.Expect(err).NotTo(HaveOccurred())
						g.Expect(output.Id).To(Equal(runConfigurationProviderId))
						g.Expect(output.ProviderError).NotTo(BeEmpty())
					}),
				Entry("Deletion fails with not found and creation succeeds", func(runConfiguration *pipelinesv1.RunConfiguration) {
					Expect(suite.SucceedUpdateNotFound(runConfiguration)).To(Succeed())
				},
					workflowFactory.ConstructUpdateWorkflow,
					func(g Gomega, workflow *argo.Workflow) {
						g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
						output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
						g.Expect(err).NotTo(HaveOccurred())
						g.Expect(output.Id).To(Equal(newRunConfigurationProviderId))
						g.Expect(output.ProviderError).To(BeEmpty())
					}),
			)

			DescribeTable("Deletion Workflow", AssertWorkflow,
				Entry("Deletion succeeds",
					func(runConfiguration *pipelinesv1.RunConfiguration) {
						Expect(suite.SucceedDeletion(runConfiguration)).To(Succeed())
					},
					workflowFactory.ConstructDeletionWorkflow,
					func(g Gomega, workflow *argo.Workflow) {
						g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
						output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
						g.Expect(err).NotTo(HaveOccurred())
						g.Expect(output.Id).To(BeEmpty())
						g.Expect(output.ProviderError).To(BeEmpty())

					},
				),
				Entry("Deletion fails",
					func(runConfiguration *pipelinesv1.RunConfiguration) {
						Expect(suite.FailDeletion(runConfiguration)).To(Succeed())
					},
					workflowFactory.ConstructDeletionWorkflow,
					func(g Gomega, workflow *argo.Workflow) {
						g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
						output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
						g.Expect(err).NotTo(HaveOccurred())
						g.Expect(output.Id).To(Equal(runConfigurationProviderId))
						g.Expect(output.ProviderError).NotTo(BeEmpty())
					},
				),
				Entry("Deletion fails with not found",
					func(runConfiguration *pipelinesv1.RunConfiguration) {
						Expect(suite.FailDeletionNotFound(runConfiguration)).To(Succeed())
					},
					workflowFactory.ConstructDeletionWorkflow,
					func(g Gomega, workflow *argo.Workflow) {
						g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
						output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
						g.Expect(err).NotTo(HaveOccurred())
						g.Expect(output.Id).To(BeEmpty())
						g.Expect(output.ProviderError).To(BeEmpty())
					},
				),
			)
		})
	}

	//RunSuite(VertexAIRunConfigurationWorkflowIntegrationSuite{}, "Vertex AI")
	RunSuite(KfpRunConfigurationWorkflowIntegrationSuite{}, "Kubeflow Pipelines")
})
