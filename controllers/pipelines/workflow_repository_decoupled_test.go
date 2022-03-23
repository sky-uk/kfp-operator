//go:build decoupled
// +build decoupled

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo/extensions/table"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"github.com/sky-uk/kfp-operator/controllers"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Context("WorkflowRepository K8s integration", func() {
	DescribeTable("Returns only non-processed workflows on retrieval", func(keepWorkflows bool) {
		optInClient := controllers.NewOptInClient(k8sManager)
		namespace := "default"

		workflowRepository := WorkflowRepositoryImpl{
			Client: optInClient,
			Config: configv1.Configuration{
				Debug: pipelinesv1.DebugOptions{
					KeepWorkflows: keepWorkflows,
				} ,
			},
		}

		operation := RandomString()
		owner := types.NamespacedName{
			Name: RandomString(),
			Namespace: namespace,
		}

		workflow := &argo.Workflow{}
		workflow.SetLabels(map[string]string{
			ExperimentWorkflowConstants.OperationLabelKey: operation,
			ExperimentWorkflowConstants.ExperimentNameLabelKey: owner.Name,
		})
		workflow.SetNamespace(namespace)
		workflow.SetName(RandomLowercaseString())

		Expect(workflowRepository.CreateWorkflow(ctx, workflow)).To(Succeed())
		Expect(workflowRepository.DeleteWorkflow(ctx, workflow)).To(Succeed())
		Expect(workflowRepository.GetByLabels(ctx, owner, map[string]string{})).To(BeEmpty())
	},
	Entry("Deletes workflows when keepWorkflows==false", false),
	Entry("Labels and filters workflows when keepWorkflows==true", true))
})

