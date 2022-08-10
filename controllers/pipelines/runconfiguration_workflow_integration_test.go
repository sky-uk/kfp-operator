//go:build integration
// +build integration

package pipelines

//
//import (
//	"fmt"
//	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
//	. "github.com/onsi/ginkgo/v2"
//	. "github.com/onsi/gomega"
//	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1alpha2"
//	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha2"
//	"github.com/walkerus/go-wiremock"
//	apiv1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/types"
//)
//
//var _ = Context("RunConfiguration Workflows", Serial, func() {
//	workflowFactory := RunConfigurationWorkflowFactory{
//		WorkflowFactoryBase: WorkflowFactoryBase{
//			Config: configv1.Configuration{
//				KfpEndpoint: "http://wiremock:80",
//				Argo: configv1.ArgoConfiguration{
//					KfpSdkImage: "kfp-operator-argo-kfp-sdk",
//					ContainerDefaults: apiv1.Container{
//						ImagePullPolicy: "Never", // Needed for minikube to use local images
//					},
//				},
//				DefaultExperiment: "Default",
//			},
//		},
//	}
//
//	pipelineKfpId := RandomString()
//	versionKfpId := RandomString()
//	versionName := RandomString()
//	jobKfpId := RandomString()
//	newJobKfpId := RandomString()
//	experimentKfpId := RandomString()
//
//	var StubGetExperiment = func(experimentName string, experimentKfpId string) error {
//		return wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/experiments")).
//			WithQueryParam("filter", wiremock.EqualToJson(
//				fmt.Sprintf(`{"predicates": [{"op": 1, "key": "name", "stringValue": "%s"}]}`, experimentName))).
//			WillReturn(
//				fmt.Sprintf(`{"experiments": [{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}]}`,
//					experimentKfpId, experimentName),
//				map[string]string{"Content-Type": "application/json"},
//				200,
//			))
//	}
//
//	var StubGetPipeline = func(pipelineName string, pipelineId string) error {
//		return wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/pipelines")).
//			WithQueryParam("filter", wiremock.EqualToJson(
//				fmt.Sprintf(`{"predicates": [{"op": 1, "key": "name", "stringValue": "%s"}]}`, pipelineName))).
//			WillReturn(
//				fmt.Sprintf(`{"pipelines": [{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}]}`,
//					pipelineId, pipelineName),
//				map[string]string{"Content-Type": "application/json"},
//				200,
//			))
//	}
//
//	var StubGetPipelineVersions = func(pipelineKfpId string) error {
//		return wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/pipeline_versions")).
//			WithQueryParam("resource_key.id", wiremock.EqualTo(pipelineKfpId)).
//			WithQueryParam("filter", wiremock.EqualToJson(
//				fmt.Sprintf(`{"predicates": [{"op": "EQUALS", "key": "name", "string_value": "%s"}]}`, versionName))).
//			WillReturn(
//				fmt.Sprintf(`{"versions": [{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}]}`,
//					versionKfpId, versionName),
//				map[string]string{"Content-Type": "application/json"},
//				200,
//			))
//	}
//
//	var SucceedCreation = func(runconfiguration *pipelinesv1.RunConfiguration, jobKfpId string) error {
//		if err := StubGetExperiment(workflowFactory.Config.DefaultExperiment, experimentKfpId); err != nil {
//			return err
//		}
//
//		if err := StubGetPipeline(runconfiguration.Spec.Pipeline.Name, pipelineKfpId); err != nil {
//			return err
//		}
//
//		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/jobs")).
//			WillReturn(
//				fmt.Sprintf(`{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}`,
//					jobKfpId, runconfiguration.Name),
//				map[string]string{"Content-Type": "application/json"},
//				200,
//			))
//	}
//
//	var FailCreation = func() error {
//		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/jobs")).
//			WillReturn(
//				`{"status": "failed"}`,
//				map[string]string{"Content-Type": "application/json"},
//				404,
//			))
//	}
//
//	var SucceedDeletion = func(jobKfpId string) error {
//		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/jobs/"+jobKfpId)).
//			WillReturn(
//				`{"status": "deleted"}`,
//				map[string]string{"Content-Type": "application/json"},
//				200,
//			))
//	}
//
//	var FailDeletionWithCode = func(jobKfpId string, code int64) error {
//		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/jobs/"+jobKfpId)).
//			WillReturn(
//				fmt.Sprintf(`HTTP response body: {"status": "failed", "code": %d}`, code),
//				map[string]string{"Content-Type": "application/json"},
//				404,
//			))
//	}
//
//	var AssertWorkflow = func(
//		setUp func(runconfiguration *pipelinesv1.RunConfiguration),
//		constructWorkflow func(*pipelinesv1.RunConfiguration) (*argo.Workflow, error),
//		assertion func(Gomega, *argo.Workflow)) {
//
//		testCtx := NewRunConfigurationTestContext(
//			&pipelinesv1.RunConfiguration{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:      RandomLowercaseString(),
//					Namespace: "argo",
//				},
//				Spec: pipelinesv1.RunConfigurationSpec{
//					Pipeline: pipelinesv1.PipelineIdentifier{Name: "pipeline", Version: versionName},
//					Schedule: "* * * * * *",
//				},
//				Status: pipelinesv1.RunConfigurationStatus{
//					Status: pipelinesv1.Status{
//						KfpId: jobKfpId,
//					},
//				},
//			},
//			k8sClient, ctx)
//
//		setUp(testCtx.RunConfiguration)
//		workflow, err := constructWorkflow(testCtx.RunConfiguration)
//
//		Expect(err).NotTo(HaveOccurred())
//		Expect(k8sClient.Create(ctx, workflow)).To(Succeed())
//
//		Eventually(testCtx.WorkflowByNameToMatch(types.NamespacedName{Name: workflow.Name, Namespace: workflow.Namespace},
//			assertion), TestTimeout).Should(Succeed())
//	}
//
//	DescribeTable("Creation Workflow", AssertWorkflow,
//		Entry("Creation succeeds",
//			func(runconfiguration *pipelinesv1.RunConfiguration) {
//				Expect(StubGetPipelineVersions(pipelineKfpId)).To(Succeed())
//				Expect(SucceedCreation(runconfiguration, jobKfpId)).To(Succeed())
//			},
//			workflowFactory.ConstructCreationWorkflow,
//			func(g Gomega, workflow *argo.Workflow) {
//				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
//				g.Expect(getWorkflowOutput(workflow, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)).
//					To(Equal(jobKfpId))
//			},
//		),
//		Entry("Creation fails",
//			func(runconfiguration *pipelinesv1.RunConfiguration) {
//				Expect(FailCreation()).To(Succeed())
//			},
//			workflowFactory.ConstructCreationWorkflow,
//			func(g Gomega, workflow *argo.Workflow) {
//				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
//			},
//		),
//	)
//
//	DescribeTable("Deletion Workflow", AssertWorkflow,
//		Entry("Deletion succeeds",
//			func(runconfiguration *pipelinesv1.RunConfiguration) {
//				Expect(SucceedDeletion(jobKfpId)).To(Succeed())
//			},
//			workflowFactory.ConstructDeletionWorkflow,
//			func(g Gomega, workflow *argo.Workflow) {
//				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
//			},
//		),
//		Entry("Deletion fails",
//			func(runconfiguration *pipelinesv1.RunConfiguration) {
//				Expect(FailDeletionWithCode(jobKfpId, 7)).To(Succeed())
//			},
//			workflowFactory.ConstructDeletionWorkflow,
//			func(g Gomega, workflow *argo.Workflow) {
//				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
//			},
//		),
//		Entry("Deletion fails with RC not found",
//			func(runconfiguration *pipelinesv1.RunConfiguration) {
//				Expect(FailDeletionWithCode(jobKfpId, 5)).To(Succeed())
//			},
//			workflowFactory.ConstructDeletionWorkflow,
//			func(g Gomega, workflow *argo.Workflow) {
//				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
//			},
//		),
//	)
//
//	DescribeTable("Update Workflow", AssertWorkflow,
//		Entry("Deletion and creation succeed", func(runconfiguration *pipelinesv1.RunConfiguration) {
//			Expect(SucceedDeletion(jobKfpId)).To(Succeed())
//			Expect(StubGetPipelineVersions(pipelineKfpId)).To(Succeed())
//			Expect(SucceedCreation(runconfiguration, newJobKfpId)).To(Succeed())
//		},
//			workflowFactory.ConstructUpdateWorkflow,
//			func(g Gomega, workflow *argo.Workflow) {
//				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
//				g.Expect(getWorkflowOutput(workflow, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)).
//					To(Equal(newJobKfpId))
//			}),
//		Entry("Deletion succeeds and creation fails", func(runconfiguration *pipelinesv1.RunConfiguration) {
//			Expect(SucceedDeletion(jobKfpId)).To(Succeed())
//			Expect(FailCreation()).To(Succeed())
//		},
//			workflowFactory.ConstructUpdateWorkflow,
//			func(g Gomega, workflow *argo.Workflow) {
//				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
//				g.Expect(getWorkflowOutput(workflow, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)).
//					To(Equal(""))
//			}),
//		Entry("Deletion fails", func(runconfiguration *pipelinesv1.RunConfiguration) {
//			Expect(FailDeletionWithCode(jobKfpId, 55)).To(Succeed())
//		},
//			workflowFactory.ConstructUpdateWorkflow,
//			func(g Gomega, workflow *argo.Workflow) {
//				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
//			}),
//		Entry("Deletion fails with RC not found and creation succeeds", func(runconfiguration *pipelinesv1.RunConfiguration) {
//			Expect(StubGetPipelineVersions(pipelineKfpId)).To(Succeed())
//			Expect(FailDeletionWithCode(jobKfpId, 5)).To(Succeed())
//			Expect(SucceedCreation(runconfiguration, newJobKfpId)).To(Succeed())
//		},
//			workflowFactory.ConstructUpdateWorkflow,
//			func(g Gomega, workflow *argo.Workflow) {
//				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
//				g.Expect(getWorkflowOutput(workflow, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName)).
//					To(Equal(newJobKfpId))
//			}),
//	)
//})
