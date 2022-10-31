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
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"github.com/walkerus/go-wiremock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Context("RunConfiguration Workflows", Serial, func() {
	workflowFactory := RunConfigurationWorkflowFactory{
		WorkflowFactoryBase: WorkflowFactoryBase{
			Config: config.Configuration{
				WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
				DefaultExperiment:      "Default",
			},
			ProviderConfig: "endpoint: http://wiremock:80\nimage: kfp-operator-kfp-provider",
		},
	}

	pipelineProviderId := apis.RandomString()
	versionProviderId := apis.RandomString()
	versionName := apis.RandomString()
	runConfigurationProviderId := apis.RandomString()
	newRunConfigurationProviderId := apis.RandomString()
	experimentProviderId := apis.RandomString()

	var StubGetExperiment = func(experimentName string, experimentProviderId string) error {
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

	var StubGetPipeline = func(pipelineName string, pipelineId string) error {
		return wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines")).
			WithQueryParam("filter", wiremock.EqualToJson(
				fmt.Sprintf(`{"predicates": [{"op": 1, "key": "name", "stringValue": "%s"}]}`, pipelineName))).
			WillReturn(
				fmt.Sprintf(`{"pipelines": [{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}]}`,
					pipelineId, pipelineName),
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var StubGetPipelineVersions = func(pipelineProviderId string) error {
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

	var SucceedCreation = func(runconfiguration *pipelinesv1.RunConfiguration, runConfigurationProviderId string) error {
		if err := StubGetExperiment(workflowFactory.Config.DefaultExperiment, experimentProviderId); err != nil {
			return err
		}

		if err := StubGetPipeline(runconfiguration.Spec.Pipeline.Name, pipelineProviderId); err != nil {
			return err
		}

		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/jobs")).
			WillReturn(
				fmt.Sprintf(`{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}`,
					runConfigurationProviderId, runconfiguration.Name),
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var FailCreation = func() error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/jobs")).
			WillReturn(
				`{"status": "failed"}`,
				map[string]string{"Content-Type": "application/json"},
				404,
			))
	}

	var SucceedDeletion = func(runConfigurationProviderId string) error {
		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/jobs/"+runConfigurationProviderId)).
			WillReturn(
				`{"status": "deleted"}`,
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var FailDeletionWithCode = func(runConfigurationProviderId string, code int64) error {
		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/jobs/"+runConfigurationProviderId)).
			WillReturn(
				fmt.Sprintf(`HTTP response body: {"status": "failed", "code": %d}`, code),
				map[string]string{"Content-Type": "application/json"},
				404,
			))
	}

	var AssertWorkflow = func(
		setUp func(runconfiguration *pipelinesv1.RunConfiguration),
		constructWorkflow func(*pipelinesv1.RunConfiguration) (*argo.Workflow, error),
		assertion func(Gomega, *argo.Workflow)) {

		testCtx := NewRunConfigurationTestContext(
			&pipelinesv1.RunConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apis.RandomLowercaseString(),
					Namespace: "argo",
				},
				Spec: pipelinesv1.RunConfigurationSpec{
					Pipeline: pipelinesv1.PipelineIdentifier{Name: "pipeline", Version: versionName},
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

	DescribeTable("Creation Workflow", AssertWorkflow,
		Entry("Creation succeeds",
			func(runconfiguration *pipelinesv1.RunConfiguration) {
				Expect(StubGetPipelineVersions(pipelineProviderId)).To(Succeed())
				Expect(SucceedCreation(runconfiguration, runConfigurationProviderId)).To(Succeed())
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
			func(runconfiguration *pipelinesv1.RunConfiguration) {
				Expect(FailCreation()).To(Succeed())
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

	DescribeTable("Deletion Workflow", AssertWorkflow,
		Entry("Deletion succeeds",
			func(runconfiguration *pipelinesv1.RunConfiguration) {
				Expect(SucceedDeletion(runConfigurationProviderId)).To(Succeed())
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
			func(runconfiguration *pipelinesv1.RunConfiguration) {
				Expect(FailDeletionWithCode(runConfigurationProviderId, 7)).To(Succeed())
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
			func(runconfiguration *pipelinesv1.RunConfiguration) {
				Expect(FailDeletionWithCode(runConfigurationProviderId, 5)).To(Succeed())
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

	DescribeTable("Update Workflow", AssertWorkflow,
		Entry("Deletion and creation succeed", func(runconfiguration *pipelinesv1.RunConfiguration) {
			Expect(SucceedDeletion(runConfigurationProviderId)).To(Succeed())
			Expect(StubGetPipelineVersions(pipelineProviderId)).To(Succeed())
			Expect(SucceedCreation(runconfiguration, newRunConfigurationProviderId)).To(Succeed())
		},
			workflowFactory.ConstructUpdateWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output.Id).To(Equal(newRunConfigurationProviderId))
				g.Expect(output.ProviderError).To(BeEmpty())
			}),
		Entry("Deletion succeeds and creation fails", func(runconfiguration *pipelinesv1.RunConfiguration) {
			Expect(SucceedDeletion(runConfigurationProviderId)).To(Succeed())
			Expect(FailCreation()).To(Succeed())
		},
			workflowFactory.ConstructUpdateWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output.Id).To(BeEmpty())
				g.Expect(output.ProviderError).NotTo(BeEmpty())
			}),
		Entry("Deletion fails", func(runconfiguration *pipelinesv1.RunConfiguration) {
			Expect(FailDeletionWithCode(runConfigurationProviderId, 55)).To(Succeed())
		},
			workflowFactory.ConstructUpdateWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output.Id).To(Equal(runConfigurationProviderId))
				g.Expect(output.ProviderError).NotTo(BeEmpty())
			}),
		Entry("Deletion fails with not found and creation succeeds", func(runconfiguration *pipelinesv1.RunConfiguration) {
			Expect(StubGetPipelineVersions(pipelineProviderId)).To(Succeed())
			Expect(FailDeletionWithCode(runConfigurationProviderId, 5)).To(Succeed())
			Expect(SucceedCreation(runconfiguration, newRunConfigurationProviderId)).To(Succeed())
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
})
