package controllers

import (
	"fmt"

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
	PipelineId        = "12345"
)

type TestContext struct {
	Pipeline                  pipelinesv1.Pipeline
	PipelineLookupKey         types.NamespacedName
	CreationWorkflowLookupKey types.NamespacedName
	UpdateWorkflowLookupKey   types.NamespacedName
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
		PipelineLookupKey:         types.NamespacedName{Name: pipelineName, Namespace: PipelineNamespace},
		CreationWorkflowLookupKey: types.NamespacedName{Name: "create-pipeline-" + pipelineName, Namespace: PipelineNamespace},
		UpdateWorkflowLookupKey:   types.NamespacedName{Name: "update-pipeline-" + pipelineName, Namespace: PipelineNamespace},
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

func (testCtx TestContext) pipelineToMatch(matcher func(Gomega, *pipelinesv1.Pipeline)) func(Gomega) {
	return func(g Gomega) {
		pipeline := &pipelinesv1.Pipeline{}
		g.Expect(k8sClient.Get(ctx, testCtx.PipelineLookupKey, pipeline)).To(Succeed())

		matcher(g, pipeline)
	}
}

func (testCtx TestContext) workflowToMatch(name types.NamespacedName, matcher func(Gomega, *argo.Workflow)) func(Gomega) {
	return func(g Gomega) {
		workflow := &argo.Workflow{}
		g.Expect(k8sClient.Get(ctx, name, workflow)).To(Succeed())

		matcher(g, workflow)
	}
}

func (testCtx TestContext) updateWorkflow(name types.NamespacedName, updateFunc func(*argo.Workflow)) error {
	workflow := &argo.Workflow{}

	if err := k8sClient.Get(ctx, name, workflow); err != nil {
		return err
	}

	updateFunc(workflow)

	return k8sClient.Update(ctx, workflow)
}

func (testCtx TestContext) updatePipeline(updateFunc func(*pipelinesv1.Pipeline)) error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := k8sClient.Get(ctx, testCtx.PipelineLookupKey, pipeline); err != nil {
		return err
	}

	updateFunc(pipeline)

	return k8sClient.Update(ctx, pipeline)
}

func (testCtx TestContext) updatePipelineStatus(updateFunc func(*pipelinesv1.Pipeline)) error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := k8sClient.Get(ctx, testCtx.PipelineLookupKey, pipeline); err != nil {
		return err
	}

	updateFunc(pipeline)

	return k8sClient.Status().Update(ctx, pipeline)
}

func (testCtx TestContext) pipelineCreatedWithStatus(status pipelinesv1.PipelineStatus) {
	Expect(k8sClient.Create(ctx, &testCtx.Pipeline)).To(Succeed())

	Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
		g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
	})).Should(Succeed())

	Expect(testCtx.updatePipelineStatus(func(pipeline *pipelinesv1.Pipeline) {
		pipeline.Status = status
	})).To(Succeed())
}

var _ = Describe("Pipeline controller", func() {

	When("Creation of a pipeline succeeds", func() {
		testCtx := NewTestContext("succeeding-pipeline")
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

		pipelineVersion, _ := pipelinesv1.ComputeVersion(testCtx.Pipeline.Spec)

		It("updates the SynchronizationStatus, Id and Version", func() {
			Expect(k8sClient.Create(ctx, &testCtx.Pipeline)).To(Succeed())

			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
			})).Should(Succeed())

			Eventually(testCtx.workflowToMatch(testCtx.CreationWorkflowLookupKey, func(g Gomega, workflow *argo.Workflow) {
				worklfowInputParameters := mapParams(workflow.Spec.Arguments.Parameters)
				actualConfig := make(map[interface{}]interface{})
				yaml.Unmarshal([]byte(worklfowInputParameters["config"]), actualConfig)
				g.Expect(actualConfig).To(Equal(expectedConfig))
			})).Should(Succeed())

			Expect(testCtx.updateWorkflow(testCtx.CreationWorkflowLookupKey, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setWorkflowOutput(workflow, PipelineId)
			})).To(Succeed())

			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
				g.Expect(pipeline.Status.Id).To(Equal(PipelineId))
				g.Expect(pipeline.Status.Version).To(Equal(pipelineVersion))
			})).Should(Succeed())
		})
	})

	When("Creation of a pipeline fails", func() {
		testCtx := NewTestContext("failing-pipeline")

		It("updates the SynchronizationStatus to failed", func() {
			Expect(k8sClient.Create(ctx, &testCtx.Pipeline)).To(Succeed())

			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
			})).Should(Succeed())

			Expect(testCtx.updateWorkflow(testCtx.CreationWorkflowLookupKey, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowFailed
			})).To(Succeed())

			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Failed))
			})).Should(Succeed())
		})
	})

	When("The resource is updated with no changes", func() {
		testCtx := NewTestContext("updated-pipeline-no-changes")

		It("keeps the status unchanged", func() {
			pipelineVersion, _ := pipelinesv1.ComputeVersion(testCtx.Pipeline.Spec)

			fmt.Println(testCtx.Pipeline.Spec)
			testCtx.pipelineCreatedWithStatus(pipelinesv1.PipelineStatus{
				Id:                   PipelineId,
				SynchronizationState: pipelinesv1.Succeeded,
				Version:              pipelineVersion,
			})

			Expect(testCtx.updatePipeline(func(pipeline *pipelinesv1.Pipeline) {
				// No changes
			})).To(Succeed())

			Consistently(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())
		})
	})

	When("The resource is updated with changes and the update succeeds", func() {
		testCtx := NewTestContext("updated-succeeding-pipeline")

		It("updates the SynchronizationStatus to Succeeded", func() {
			pipelineVersion, _ := pipelinesv1.ComputeVersion(testCtx.Pipeline.Spec)

			testCtx.pipelineCreatedWithStatus(pipelinesv1.PipelineStatus{
				Id:                   PipelineId,
				SynchronizationState: pipelinesv1.Succeeded,
				Version:              pipelineVersion,
			})

			modifiedSpec := testCtx.Pipeline.Spec
			modifiedSpec.Env["c"] = "cVal"
			modifiedVersion, _ := pipelinesv1.ComputeVersion(modifiedSpec)

			Expect(testCtx.updatePipeline(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = modifiedSpec
			})).To(Succeed())

			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
			})).Should(Succeed())

			Expect(testCtx.updateWorkflow(testCtx.UpdateWorkflowLookupKey, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
			})).To(Succeed())

			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
				g.Expect(pipeline.Status.Version).To(Equal(modifiedVersion))
			})).Should(Succeed())
		})
	})

	When("The resource is updated with changes and the update fails", func() {
		testCtx := NewTestContext("updated-failing-pipeline")

		It("updates the SynchronizationStatus to Succeeded", func() {
			pipelineVersion, _ := pipelinesv1.ComputeVersion(testCtx.Pipeline.Spec)

			testCtx.pipelineCreatedWithStatus(pipelinesv1.PipelineStatus{
				Id:                   PipelineId,
				SynchronizationState: pipelinesv1.Succeeded,
				Version:              pipelineVersion,
			})

			modifiedSpec := testCtx.Pipeline.Spec
			modifiedSpec.Env["c"] = "cVal"
			modifiedVersion, _ := pipelinesv1.ComputeVersion(modifiedSpec)

			Expect(testCtx.updatePipeline(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = modifiedSpec
			})).To(Succeed())

			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
			})).Should(Succeed())

			Expect(testCtx.updateWorkflow(testCtx.UpdateWorkflowLookupKey, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowFailed
			})).To(Succeed())

			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Failed))
				g.Expect(pipeline.Status.Version).To(Equal(modifiedVersion))
			})).Should(Succeed())
		})
	})
})
