//go:build decoupled
// +build decoupled

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"github.com/sky-uk/kfp-operator/controllers"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Context("WorkflowRepository K8s integration", func() {
	DescribeTable("Returns only non-processed workflows on retrieval", func(keepWorkflows bool) {
		optInClient := controllers.NewOptInClient(k8sManager)
		namespace := "default"

		scheme := runtime.NewScheme()
		pipelinesv1.AddToScheme(scheme)

		workflowRepository := WorkflowRepositoryImpl{
			Client: optInClient,
			Scheme: scheme,
			Config: configv1.Configuration{
				Debug: pipelinesv1.DebugOptions{
					KeepWorkflows: keepWorkflows,
				},
			},
		}

		owner := &pipelinesv1.Pipeline{}
		owner.SetName(RandomString())
		owner.SetUID(types.UID(RandomString()))
		workflow := &argo.Workflow{}
		workflow.SetNamespace(namespace)
		workflow.SetName(RandomLowercaseString())

		Expect(workflowRepository.CreateWorkflowForResource(ctx, workflow, owner)).To(Succeed())
		Expect(workflowRepository.DeleteWorkflow(ctx, workflow)).To(Succeed())
		Expect(workflowRepository.GetByLabels(ctx, owner.Name, map[string]string{})).To(BeEmpty())
	},
		Entry("Deletes workflows when keepWorkflows==false", false),
		Entry("Filters workflows when keepWorkflows==true", true))
})
