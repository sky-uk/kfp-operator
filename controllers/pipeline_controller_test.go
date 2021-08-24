package controllers

import (
	"context"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/api/v1"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	// +kubebuilder:scaffold:imports
)

const (
	PipelineNamespace = "default"
)

type TestContext struct {
	context.Context
	Pipeline          pipelinesv1.Pipeline
	PipelineLookupKey types.NamespacedName
	WorkflowLookupKey types.NamespacedName
}

func NewTestContext(pipelineName string) TestContext {
	return TestContext{
		Pipeline: pipelinesv1.Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pipelineName,
				Namespace: PipelineNamespace,
			},
			Spec: pipelinesv1.PipelineSpec{
				Image:         "image:v1",
				TfxComponents: "pipeline.create_components",
				Env: map[string]string{
					"a": "aVal",
					"b": "bVal",
				},
			},
		},
		PipelineLookupKey: types.NamespacedName{Name: pipelineName, Namespace: PipelineNamespace},
		WorkflowLookupKey: types.NamespacedName{Name: "create-pipeline-" + pipelineName, Namespace: PipelineNamespace},
	}
}

func setWorkflowOutput(workflow *argo.Workflow, output string) {
	result := argo.AnyString(output)
	nodes := make(map[string]argo.NodeStatus)
	nodes[workflow.ObjectMeta.Name] = argo.NodeStatus{
		Outputs: &argo.Outputs{
			Parameters: []argo.Parameter{
				{
					Value: &result,
				},
			},
		},
	}

	workflow.Status.Nodes = nodes
}

func (ct TestContext) pipelineToMatch(matcher func(Gomega, *pipelinesv1.Pipeline)) func(Gomega) {
	return func(g Gomega) {
		pipeline := &pipelinesv1.Pipeline{}
		g.Expect(k8sClient.Get(ctx, ct.PipelineLookupKey, pipeline)).To(Succeed())

		matcher(g, pipeline)
	}
}

func (ct TestContext) workflowToMatch(matcher func(Gomega, *argo.Workflow)) func(Gomega) {
	return func(g Gomega) {
		workflow := &argo.Workflow{}
		g.Expect(k8sClient.Get(ctx, ct.WorkflowLookupKey, workflow)).To(Succeed())

		matcher(g, workflow)
	}
}

func (ct TestContext) updateWorkflow(updateFunc func(*argo.Workflow)) error {
	workflow := &argo.Workflow{}

	if err := k8sClient.Get(ctx, ct.WorkflowLookupKey, workflow); err != nil {
		return err
	}

	updateFunc(workflow)

	return k8sClient.Update(ctx, workflow)
}

var _ = Describe("Pipeline controller", func() {
	When("Creation of a pipeline succeeds", func() {
		ct := NewTestContext("succeeding-pipeline")
		pipelineId := "12345"
		var mapParams = func(params []argo.Parameter) map[string]string {
			m := make(map[string]string)
			for i := range params {
				m[params[i].Name] = string(*params[i].Value)
			}

			return m
		}

		expectedConfig := map[interface{}]interface{}{
			"image":         "image:v1",
			"tfxComponents": "pipeline.create_components",
			"env": map[interface{}]interface{}{
				"a": "aVal",
				"b": "bVal",
			},
		}

		It("updates the SynchronizationStatus and Id", func() {
			Expect(k8sClient.Create(ctx, &ct.Pipeline)).Should(Succeed())

			Eventually(ct.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
			})).Should(Succeed())

			Eventually(ct.workflowToMatch(func(g Gomega, workflow *argo.Workflow) {
				params := mapParams(workflow.Spec.Arguments.Parameters)
				actualConfig := make(map[interface{}]interface{})
				yaml.Unmarshal([]byte(params["config"]), actualConfig)
				g.Expect(actualConfig).To(Equal(expectedConfig))
			})).Should(Succeed())

			Expect(ct.updateWorkflow(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setWorkflowOutput(workflow, pipelineId)
			})).To(Succeed())

			Eventually(ct.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
				g.Expect(pipeline.Status.Id).To(Equal(pipelineId))
			})).Should(Succeed())
		})
	})

	When("Creation of a pipeline fails", func() {
		ct := NewTestContext("failing-pipeline")

		It("updates the SynchronizationStatus to failed", func() {
			Expect(k8sClient.Create(ctx, &ct.Pipeline)).Should(Succeed())

			Eventually(ct.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
			})).Should(Succeed())

			Expect(ct.updateWorkflow(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowFailed
			})).To(Succeed())

			Eventually(ct.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Failed))
			})).Should(Succeed())
		})
	})
})
