//go:build integration
// +build integration

package pipelines

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"github.com/walkerus/go-wiremock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Context("RunConfiguration Workflows", func() {
	workflowFactory := RunConfigurationWorkflowFactory{
		WorkflowFactory: WorkflowFactory{
			Config: configv1.Configuration{
				KfpEndpoint:       "http://wiremock:80",
				KfpSdkImage:       "kfp-operator-argo-kfp-sdk",
				ImagePullPolicy:   "Never", // Needed for minikube to use local images
				DefaultExperiment: "Default",
			},
		},
	}

	var pipelineKfpId = "12345"

	var StubGetExperiment = func(experimentName string, experimentId string) error {
		return wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/experiments")).
			WithQueryParam("filter", wiremock.EqualTo(
				fmt.Sprintf(`{"predicates": [{"op": 1, "key": "name", "stringValue": "%s"}]}`, experimentName))).
			WillReturn(
				fmt.Sprintf(`{"experiments": [{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}]}`,
					experimentId, experimentName),
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var StubGetPipeline = func(pipelineName string, pipelineId string) error {
		return wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines")).
			WithQueryParam("filter", wiremock.EqualTo(
				fmt.Sprintf(`{"predicates": [{"op": 1, "key": "name", "stringValue": "%s"}]}`, pipelineName))).
			WillReturn(
				fmt.Sprintf(`{"experiments": [{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}]}`,
					pipelineId, pipelineName),
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var SucceedCreation = func(runconfiguration *pipelinesv1.RunConfiguration, jobId string) error {
		if err := StubGetExperiment(workflowFactory.Config.DefaultExperiment, ExperimentId); err != nil {
			return err
		}
		if err := StubGetPipeline(runconfiguration.Spec.PipelineName, pipelineKfpId); err != nil {
			return err
		}

		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/jobs")).
			WillReturn(
				fmt.Sprintf(`{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}`,
					jobId, runconfiguration.Name),
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

	var SucceedDeletion = func(jobId string) error {
		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/jobs/"+jobId)).
			WillReturn(
				`{"status": "deleted"}`,
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var FailDeletion = func(jobId string) error {
		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/jobs/"+jobId)).
			WillReturn(
				`{"status": "failed"}`,
				map[string]string{"Content-Type": "application/json"},
				404,
			))
	}

	var AssertWorkflow = func(
		setUp func(runconfiguration *pipelinesv1.RunConfiguration),
		constructWorkflow func(*pipelinesv1.RunConfiguration) *argo.Workflow,
		assertion func(Gomega, *argo.Workflow)) {

		testCtx := NewRunConfigurationTestContext(
			&pipelinesv1.RunConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name:      RandomLowercaseString(),
					Namespace: "argo",
				},
				Spec: pipelinesv1.RunConfigurationSpec{
					PipelineName: "pipeline",
					Schedule:     "* * * * * *",
				},
				Status: pipelinesv1.Status{
					KfpId: JobId,
				},
			},
			k8sClient, ctx)

		setUp(testCtx.RunConfiguration)
		workflow := constructWorkflow(testCtx.RunConfiguration)

		Expect(k8sClient.Create(ctx, workflow)).To(Succeed())

		Eventually(testCtx.WorkflowByNameToMatch(types.NamespacedName{Name: workflow.Name, Namespace: workflow.Namespace},
			assertion), TestTimeout).Should(Succeed())
	}

	DescribeTable("Creation Workflow", AssertWorkflow,
		Entry("Creation succeeds",
			func(runconfiguration *pipelinesv1.RunConfiguration) {
				Expect(SucceedCreation(runconfiguration, JobId)).To(Succeed())
			},
			workflowFactory.ConstructCreationWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				g.Expect(getWorkflowOutput(workflow, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)).
					To(Equal(JobId))
			},
		),
		Entry("Creation fails",
			func(runconfiguration *pipelinesv1.RunConfiguration) {
				Expect(FailCreation()).To(Succeed())
			},
			workflowFactory.ConstructCreationWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
			},
		),
	)

	DescribeTable("Deletion Workflow", AssertWorkflow,
		Entry("Deletion succeeds",
			func(runconfiguration *pipelinesv1.RunConfiguration) {
				Expect(SucceedDeletion(JobId)).To(Succeed())
			},
			workflowFactory.ConstructDeletionWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
			},
		),
		Entry("Deletion fails",
			func(runconfiguration *pipelinesv1.RunConfiguration) {
				Expect(FailDeletion(JobId)).To(Succeed())
			},
			workflowFactory.ConstructDeletionWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
			},
		),
	)

	DescribeTable("Update Workflow", AssertWorkflow,
		Entry("Deletion and creation succeed", func(runconfiguration *pipelinesv1.RunConfiguration) {
			Expect(SucceedDeletion(JobId)).To(Succeed())
			Expect(SucceedCreation(runconfiguration, NewJobId)).To(Succeed())
		},
			workflowFactory.ConstructUpdateWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				g.Expect(getWorkflowOutput(workflow, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)).
					To(Equal(NewJobId))
			}),
		Entry("Deletion fails and creation succeeds", func(runconfiguration *pipelinesv1.RunConfiguration) {
			Expect(FailDeletion(JobId)).To(Succeed())
			Expect(SucceedCreation(runconfiguration, NewJobId)).To(Succeed())
		},
			workflowFactory.ConstructUpdateWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				g.Expect(getWorkflowOutput(workflow, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)).
					To(Equal(NewJobId))
			}),
		Entry("Deletion succeeds and creation fails", func(runconfiguration *pipelinesv1.RunConfiguration) {
			Expect(SucceedDeletion(JobId)).To(Succeed())
			Expect(FailCreation()).To(Succeed())
		},
			workflowFactory.ConstructUpdateWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
			}),
		Entry("Deletion and creation fail", func(runconfiguration *pipelinesv1.RunConfiguration) {
			Expect(FailDeletion(JobId)).To(Succeed())
			Expect(FailCreation()).To(Succeed())
		},
			workflowFactory.ConstructUpdateWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
			}),
	)
})
