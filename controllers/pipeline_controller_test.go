package controllers

import (
	"context"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	// +kubebuilder:scaffold:imports
)

var _ = Describe("Pipeline controller", func() {
	ctx := context.Background()

	const (
		PipelineName      = "my-pipeline"
		PipelineNamespace = "default"
		WorkflowName      = "create-pipeline-" + PipelineName

		PipelineId = "12345"
	)

	var (
		pipelineLookupKey = types.NamespacedName{Name: PipelineName, Namespace: PipelineNamespace}
		workflowLookupKey = types.NamespacedName{Name: WorkflowName, Namespace: PipelineNamespace}
	)

	var pipeline = func() *pipelinesv1.Pipeline {
		return &pipelinesv1.Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name:      PipelineName,
				Namespace: PipelineNamespace,
			},
			Spec: pipelinesv1.PipelineSpec{},
		}
	}

	var setWorkflowOutput = func(workflow *argo.Workflow, output string) {
		result := argo.AnyString(output)
		nodes := make(map[string]argo.NodeStatus)
		nodes[workflow.ObjectMeta.Name] = argo.NodeStatus{
			Outputs: &argo.Outputs{
				Parameters: []argo.Parameter{
					argo.Parameter{
						Value: &result,
					},
				},
			},
		}

		workflow.Status.Nodes = nodes
	}

	var pipelineToMatch = func(matcher func(Gomega, *pipelinesv1.Pipeline)) func(Gomega) {
		return func(g Gomega) {
			pipeline := &pipelinesv1.Pipeline{}
			g.Expect(k8sClient.Get(ctx, pipelineLookupKey, pipeline)).To(Succeed())

			matcher(g, pipeline)
		}
	}

	var workflowToMatch = func(matcher func(Gomega, *argo.Workflow)) func(Gomega) {
		return func(g Gomega) {
			workflow := &argo.Workflow{}
			g.Expect(k8sClient.Get(ctx, workflowLookupKey, workflow)).To(Succeed())

			matcher(g, workflow)
		}
	}

	var updateWorkflow = func(updateFunc func(*argo.Workflow)) error {
		workflow := &argo.Workflow{}

		if err := k8sClient.Get(ctx, workflowLookupKey, workflow); err != nil {
			return err
		}

		updateFunc(workflow)

		return k8sClient.Update(ctx, workflow)
	}

	When("Creation of a pipeline resource succeeds", func() {
		It("updates the SynchronizationStatus and Id", func() {
			Expect(k8sClient.Create(ctx, pipeline())).Should(Succeed())

			Eventually(pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
			})).Should(Succeed())

			Eventually(workflowToMatch(func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.ObjectMeta.Name).To(Equal(WorkflowName))
				//TODO implement workflow input check
			})).Should(Succeed())

			Expect(updateWorkflow(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setWorkflowOutput(workflow, PipelineId)
			})).To(Succeed())

			Eventually(pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
				g.Expect(pipeline.Status.Id).To(Equal(PipelineId))
			})).Should(Succeed())
		})
	})
})
