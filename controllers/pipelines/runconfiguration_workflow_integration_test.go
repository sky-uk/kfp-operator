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

var _ = Context("RunConfiguration Workflows", func() {
	workflowFactory := RunConfigurationWorkflowFactory{
		WorkflowFactory: WorkflowFactory{
			Config: configv1.Configuration{
				KfpEndpoint: "http://wiremock:80",
				Argo: configv1.ArgoConfiguration{
					KfpSdkImage: "kfp-operator-argo-kfp-sdk",
					ContainerDefaults: apiv1.Container{
						ImagePullPolicy: "Never", // Needed for minikube to use local images
					},
				},
				DefaultExperiment: "Default",
			},
		},
	}

	pipelineKfpId := RandomString()
	jobKfpId := RandomString()
	newJobKfpId := RandomString()
	experimentKfpId := RandomString()

	var StubGetExperiment = func(experimentName string, experimentKfpId string) error {
		return wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/experiments")).
			WithQueryParam("filter", wiremock.EqualTo(
				fmt.Sprintf(`{"predicates": [{"op": 1, "key": "name", "stringValue": "%s"}]}`, experimentName))).
			WillReturn(
				fmt.Sprintf(`{"experiments": [{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}]}`,
					experimentKfpId, experimentName),
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

	var SucceedCreation = func(runconfiguration *pipelinesv1.RunConfiguration, jobKfpId string) error {
		if err := StubGetExperiment(workflowFactory.Config.DefaultExperiment, experimentKfpId); err != nil {
			return err
		}
		if err := StubGetPipeline(runconfiguration.Spec.PipelineName, pipelineKfpId); err != nil {
			return err
		}

		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/jobs")).
			WillReturn(
				fmt.Sprintf(`{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}`,
					jobKfpId, runconfiguration.Name),
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

	var SucceedDeletion = func(jobKfpId string) error {
		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/jobs/"+jobKfpId)).
			WillReturn(
				`{"status": "deleted"}`,
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var FailDeletion = func(jobKfpId string) error {
		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/jobs/"+jobKfpId)).
			WillReturn(
				`{"status": "failed"}`,
				map[string]string{"Content-Type": "application/json"},
				404,
			))
	}

	var AssertWorkflow = func(
		setUp func(runconfiguration *pipelinesv1.RunConfiguration),
		constructWorkflow func(context.Context, *pipelinesv1.RunConfiguration) *argo.Workflow,
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
					KfpId: jobKfpId,
				},
			},
			k8sClient, ctx)

		setUp(testCtx.RunConfiguration)
		workflow := constructWorkflow(testCtx.ctx, testCtx.RunConfiguration)

		Expect(k8sClient.Create(ctx, workflow)).To(Succeed())

		Eventually(testCtx.WorkflowByNameToMatch(types.NamespacedName{Name: workflow.Name, Namespace: workflow.Namespace},
			assertion), TestTimeout).Should(Succeed())
	}

	DescribeTable("Creation Workflow", AssertWorkflow,
		Entry("Creation succeeds",
			func(runconfiguration *pipelinesv1.RunConfiguration) {
				Expect(SucceedCreation(runconfiguration, jobKfpId)).To(Succeed())
			},
			workflowFactory.ConstructCreationWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				g.Expect(getWorkflowOutput(workflow, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)).
					To(Equal(jobKfpId))
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
				Expect(SucceedDeletion(jobKfpId)).To(Succeed())
			},
			workflowFactory.ConstructDeletionWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
			},
		),
		Entry("Deletion fails",
			func(runconfiguration *pipelinesv1.RunConfiguration) {
				Expect(FailDeletion(jobKfpId)).To(Succeed())
			},
			workflowFactory.ConstructDeletionWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
			},
		),
	)

	DescribeTable("Update Workflow", AssertWorkflow,
		Entry("Deletion and creation succeed", func(runconfiguration *pipelinesv1.RunConfiguration) {
			Expect(SucceedDeletion(jobKfpId)).To(Succeed())
			Expect(SucceedCreation(runconfiguration, newJobKfpId)).To(Succeed())
		},
			workflowFactory.ConstructUpdateWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				g.Expect(getWorkflowOutput(workflow, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)).
					To(Equal(newJobKfpId))
			}),
		Entry("Deletion fails and creation succeeds", func(runconfiguration *pipelinesv1.RunConfiguration) {
			Expect(FailDeletion(jobKfpId)).To(Succeed())
			Expect(SucceedCreation(runconfiguration, newJobKfpId)).To(Succeed())
		},
			workflowFactory.ConstructUpdateWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				g.Expect(getWorkflowOutput(workflow, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)).
					To(Equal(newJobKfpId))
			}),
		Entry("Deletion succeeds and creation fails", func(runconfiguration *pipelinesv1.RunConfiguration) {
			Expect(SucceedDeletion(jobKfpId)).To(Succeed())
			Expect(FailCreation()).To(Succeed())
		},
			workflowFactory.ConstructUpdateWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
			}),
		Entry("Deletion and creation fail", func(runconfiguration *pipelinesv1.RunConfiguration) {
			Expect(FailDeletion(jobKfpId)).To(Succeed())
			Expect(FailCreation()).To(Succeed())
		},
			workflowFactory.ConstructUpdateWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
			}),
	)
})
