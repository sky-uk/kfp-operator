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

var _ = Context("Experiment Workflows", Serial, func() {
	workflowFactory := ExperimentWorkflowFactory{
		WorkflowFactoryBase: WorkflowFactoryBase{
			Config: config.Configuration{
				DefaultExperiment:      "Default",
				DefaultProvider:        "kfp",
				WorkflowTemplatePrefix: "kfp-operator-integration-tests-", // Needs to match integration-test-values.yaml
				WorkflowNamespace:      "argo",
			},
		},
	}

	experimentProviderId := apis.RandomString()
	newExperimentProviderId := apis.RandomString()

	var NoExperimentExists = func() error {
		return wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/experiments")).
			WillReturn(
				fmt.Sprintf(`{}`),
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var SucceedCreation = func(experiment *pipelinesv1.Experiment, experimentProviderId string) error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/experiments")).
			WillReturn(
				fmt.Sprintf(`{"id": "%s", "created_at": "2021-09-10T15:46:08Z", "name": "%s"}`,
					experimentProviderId, experiment.Name),
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var FailCreation = func() error {
		return wiremockClient.StubFor(wiremock.Post(wiremock.URLPathEqualTo("/apis/v1beta1/experiments")).
			WillReturn(
				`{"status": "failed"}`,
				map[string]string{"Content-Type": "application/json"},
				404,
			))
	}

	var SucceedDeletion = func(experimentProviderId string) error {
		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/experiments/"+experimentProviderId)).
			WillReturn(
				`{"status": "deleted"}`,
				map[string]string{"Content-Type": "application/json"},
				200,
			))
	}

	var FailDeletion = func(experimentProviderId string) error {
		return wiremockClient.StubFor(wiremock.Delete(wiremock.URLPathEqualTo("/apis/v1beta1/experiments/"+experimentProviderId)).
			WillReturn(
				`{"status": "failed"}`,
				map[string]string{"Content-Type": "application/json"},
				404,
			))
	}

	var AssertWorkflow = func(
		setUp func(experiment *pipelinesv1.Experiment),
		constructWorkflow func(*pipelinesv1.Experiment) (*argo.Workflow, error),
		assertion func(Gomega, *argo.Workflow)) {

		testCtx := NewExperimentTestContext(
			&pipelinesv1.Experiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apis.RandomLowercaseString(),
					Namespace: "argo",
				},
				Spec: pipelinesv1.ExperimentSpec{
					Description: "a description",
				},
				Status: pipelinesv1.Status{
					ProviderId: experimentProviderId,
				},
			},
			k8sClient, ctx)

		setUp(testCtx.Experiment)
		workflow, err := constructWorkflow(testCtx.Experiment)

		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient.Create(ctx, workflow)).To(Succeed())

		Eventually(testCtx.WorkflowByNameToMatch(types.NamespacedName{Name: workflow.Name, Namespace: workflow.Namespace},
			assertion), TestTimeout).Should(Succeed())
	}

	DescribeTable("Creation Workflow", AssertWorkflow,
		Entry("Creation succeeds",
			func(experiment *pipelinesv1.Experiment) {
				Expect(NoExperimentExists()).To(Succeed())
				Expect(SucceedCreation(experiment, experimentProviderId)).To(Succeed())
			},
			workflowFactory.ConstructCreationWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output.Id).To(Equal(experimentProviderId))
				g.Expect(output.ProviderError).To(BeEmpty())
			},
		),
		Entry("Creation fails",
			func(experiment *pipelinesv1.Experiment) {
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
			func(experiment *pipelinesv1.Experiment) {
				Expect(SucceedDeletion(experimentProviderId)).To(Succeed())
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
			func(experiment *pipelinesv1.Experiment) {
				Expect(FailDeletion(experimentProviderId)).To(Succeed())
			},
			workflowFactory.ConstructDeletionWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output.Id).To(Equal(experimentProviderId))
				g.Expect(output.ProviderError).NotTo(BeEmpty())
			},
		),
	)

	DescribeTable("Update Workflow", AssertWorkflow,
		Entry("Deletion and creation succeed", func(experiment *pipelinesv1.Experiment) {
			Expect(SucceedDeletion(experimentProviderId)).To(Succeed())
			Expect(NoExperimentExists()).To(Succeed())
			Expect(SucceedCreation(experiment, newExperimentProviderId)).To(Succeed())
		},
			workflowFactory.ConstructUpdateWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output.Id).To(Equal(newExperimentProviderId))
				g.Expect(output.ProviderError).To(BeEmpty())
			}),
		Entry("Deletion succeeds and creation fails", func(experiment *pipelinesv1.Experiment) {
			Expect(SucceedDeletion(experimentProviderId)).To(Succeed())
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
		Entry("Deletion fails", func(experiment *pipelinesv1.Experiment) {
			Expect(FailDeletion(experimentProviderId)).To(Succeed())
		},
			workflowFactory.ConstructUpdateWorkflow,
			func(g Gomega, workflow *argo.Workflow) {
				output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output.Id).To(Equal(experimentProviderId))
				g.Expect(output.ProviderError).NotTo(BeEmpty())
			}),
	)
})
