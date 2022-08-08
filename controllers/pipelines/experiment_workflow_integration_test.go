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
//var _ = Context("Experiment Workflows", Serial, func() {
//	workflowFactory := ExperimentWorkflowFactory{
//		WorkflowFactory: WorkflowFactory{
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
//	experimentKfpId := RandomString()
//	newExperimentKfpId := RandomString()
//
//	var NoExperimentExists = func() error {
//		return wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/experiments")).
//			WillReturn(
//				fmt.Sprintf(`{}`),
//				map[string]string{"Content-Type": "application/json"},
//				200,
//			))
//	}
//
//	var SucceedCreation = func(experiment *pipelinesv1.Experiment, experimentKfpId string) error {
//		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/experiments")).
//			WillReturn(
//				fmt.Sprintf(`{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}`,
//					experimentKfpId, experiment.Name),
//				map[string]string{"Content-Type": "application/json"},
//				200,
//			))
//	}
//
//	var FailCreation = func() error {
//		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/experiments")).
//			WillReturn(
//				`{"status": "failed"}`,
//				map[string]string{"Content-Type": "application/json"},
//				404,
//			))
//	}
//
//	var SucceedDeletion = func(experimentKfpId string) error {
//		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/experiments/"+experimentKfpId)).
//			WillReturn(
//				`{"status": "deleted"}`,
//				map[string]string{"Content-Type": "application/json"},
//				200,
//			))
//	}
//
//	var FailDeletion = func(experimentKfpId string) error {
//		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/experiments/"+experimentKfpId)).
//			WillReturn(
//				`{"status": "failed"}`,
//				map[string]string{"Content-Type": "application/json"},
//				404,
//			))
//	}
//
//	var AssertWorkflow = func(
//		setUp func(experiment *pipelinesv1.Experiment),
//		constructWorkflow func(*pipelinesv1.Experiment) (*argo.Workflow, error),
//		assertion func(Gomega, *argo.Workflow)) {
//
//		testCtx := NewExperimentTestContext(
//			&pipelinesv1.Experiment{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:      RandomLowercaseString(),
//					Namespace: "argo",
//				},
//				Spec: pipelinesv1.ExperimentSpec{
//					Description: "a description",
//				},
//				Status: pipelinesv1.Status{
//					KfpId: experimentKfpId,
//				},
//			},
//			k8sClient, ctx)
//
//		setUp(testCtx.Experiment)
//		workflow, err := constructWorkflow(testCtx.Experiment)
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
//			func(experiment *pipelinesv1.Experiment) {
//				Expect(NoExperimentExists()).To(Succeed())
//				Expect(SucceedCreation(experiment, experimentKfpId)).To(Succeed())
//			},
//			workflowFactory.ConstructCreationWorkflow,
//			func(g Gomega, workflow *argo.Workflow) {
//				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
//				g.Expect(getWorkflowOutput(workflow, ExperimentWorkflowConstants.ExperimentIdParameterName)).
//					To(Equal(experimentKfpId))
//			},
//		),
//		Entry("Creation fails",
//			func(experiment *pipelinesv1.Experiment) {
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
//			func(experiment *pipelinesv1.Experiment) {
//				Expect(SucceedDeletion(experimentKfpId)).To(Succeed())
//			},
//			workflowFactory.ConstructDeletionWorkflow,
//			func(g Gomega, workflow *argo.Workflow) {
//				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
//			},
//		),
//		Entry("Deletion fails",
//			func(experiment *pipelinesv1.Experiment) {
//				Expect(FailDeletion(experimentKfpId)).To(Succeed())
//			},
//			workflowFactory.ConstructDeletionWorkflow,
//			func(g Gomega, workflow *argo.Workflow) {
//				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
//			},
//		),
//	)
//
//	DescribeTable("Update Workflow", AssertWorkflow,
//		Entry("Deletion and creation succeed", func(experiment *pipelinesv1.Experiment) {
//			Expect(SucceedDeletion(experimentKfpId)).To(Succeed())
//			Expect(NoExperimentExists()).To(Succeed())
//			Expect(SucceedCreation(experiment, newExperimentKfpId)).To(Succeed())
//		},
//			workflowFactory.ConstructUpdateWorkflow,
//			func(g Gomega, workflow *argo.Workflow) {
//				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
//				g.Expect(getWorkflowOutput(workflow, ExperimentWorkflowConstants.ExperimentIdParameterName)).
//					To(Equal(newExperimentKfpId))
//			}),
//		Entry("Deletion succeeds and creation fails", func(experiment *pipelinesv1.Experiment) {
//			Expect(SucceedDeletion(experimentKfpId)).To(Succeed())
//			Expect(FailCreation()).To(Succeed())
//		},
//			workflowFactory.ConstructUpdateWorkflow,
//			func(g Gomega, workflow *argo.Workflow) {
//				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
//				g.Expect(getWorkflowOutput(workflow, ExperimentWorkflowConstants.ExperimentIdParameterName)).
//					To(Equal(""))
//			}),
//		Entry("Deletion fails", func(experiment *pipelinesv1.Experiment) {
//			Expect(FailDeletion(experimentKfpId)).To(Succeed())
//		},
//			workflowFactory.ConstructUpdateWorkflow,
//			func(g Gomega, workflow *argo.Workflow) {
//				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowFailed))
//			}),
//	)
//})
