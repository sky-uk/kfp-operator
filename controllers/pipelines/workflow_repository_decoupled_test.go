//go:build decoupled

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers"
	. "github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const namespace = "default"

func createWorkflowRepository() WorkflowRepositoryImpl {
	optInClient := controllers.NewOptInClient(k8sManager)

	scheme := runtime.NewScheme()
	pipelinesv1.AddToScheme(scheme)

	return WorkflowRepositoryImpl{
		Client: optInClient,
		Scheme: scheme,
	}
}

func randomResource() pipelinesv1.Resource {
	resource := &pipelinesv1.Pipeline{}
	resource.SetName(apis.RandomString())
	resource.SetUID(types.UID(apis.RandomString()))

	return resource
}

func randomWorkflow() *argo.Workflow {
	workflow := &argo.Workflow{}
	workflow.SetNamespace(namespace)
	workflow.SetName(apis.RandomLowercaseString())

	randomLabels := map[string]string{
		apis.RandomString(): apis.RandomString(),
	}
	workflow.SetLabels(randomLabels)

	return workflow
}

var _ = Context("WorkflowRepository K8s integration", Serial, func() {
	_ = Describe("GetByLabels", func() {
		It("Returns only non-processed workflows on retrieval", func() {
			workflowRepository := createWorkflowRepository()

			owner := randomResource()
			workflow := randomWorkflow()

			Expect(workflowRepository.CreateWorkflowForResource(Ctx, workflow, owner)).To(Succeed())
			Expect(workflowRepository.GetByLabels(Ctx, workflow.GetLabels())).To(Not(BeEmpty()))
			Expect(workflowRepository.MarkWorkflowAsProcessed(Ctx, workflow)).To(Succeed())
			Expect(workflowRepository.GetByLabels(Ctx, workflow.GetLabels())).To(BeEmpty())
		})
	})
})
