//go:build decoupled
// +build decoupled

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1alpha3"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	"github.com/sky-uk/kfp-operator/controllers"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const namespace = "default"

func createWorkflowRepository(keepWorkflows bool) WorkflowRepositoryImpl {
	optInClient := controllers.NewOptInClient(k8sManager)

	scheme := runtime.NewScheme()
	pipelinesv1.AddToScheme(scheme)

	return WorkflowRepositoryImpl{
		Client: optInClient,
		Scheme: scheme,
		Config: configv1.Configuration{
			Debug: apis.DebugOptions{
				KeepWorkflows: keepWorkflows,
			},
		},
	}
}

func randomResource() Resource {
	resource := &pipelinesv1.Pipeline{}
	resource.SetName(RandomString())
	resource.SetUID(types.UID(RandomString()))

	return resource
}

func randomWorkflow() *argo.Workflow {
	workflow := &argo.Workflow{}
	workflow.SetNamespace(namespace)
	workflow.SetName(RandomLowercaseString())

	randomLabels := map[string]string{
		RandomString(): RandomString(),
	}
	workflow.SetLabels(randomLabels)

	return workflow
}

var _ = Context("WorkflowRepository K8s integration", Serial, func() {
	_ = Describe("Creating Workflows", func() {
		It("Sets ownership", func() {
			workflowRepository := createWorkflowRepository(false)

			owner := randomResource()
			workflow := randomWorkflow()

			Expect(workflowRepository.CreateWorkflowForResource(ctx, workflow, owner)).To(Succeed())
			retrievedWorkflows := workflowRepository.GetByLabels(ctx, namespace, workflow.GetLabels())
			Expect(retrievedWorkflows[0].GetOwnerReferences()[0].UID).To(Equal(owner.GetUID()))
		})
	})

	DescribeTable("Returns only non-processed workflows on retrieval", func(keepWorkflows bool) {
		workflowRepository := createWorkflowRepository(keepWorkflows)

		owner := randomResource()
		workflow := randomWorkflow()

		Expect(workflowRepository.CreateWorkflowForResource(ctx, workflow, owner)).To(Succeed())
		Expect(workflowRepository.GetByLabels(ctx, namespace, workflow.GetLabels())).To(Not(BeEmpty()))
		Expect(workflowRepository.DeleteWorkflow(ctx, workflow)).To(Succeed())
		Expect(workflowRepository.GetByLabels(ctx, namespace, workflow.GetLabels())).To(BeEmpty())
	},
		Entry("keepWorkflows is disabled", false),
		Entry("keepWorkflows is enabled", true),
	)
})
