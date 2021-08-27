package controllers

import (
	"fmt"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/api/v1"
	"github.com/thanhpk/randstr"
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
	DeletionWorkflowLookupKey types.NamespacedName
	Version                   string
}

func NewTestContext() TestContext {
	pipelineName := randstr.String(16, "0123456789abcdefghijklmnopqrstuvwxyz")
	pipeline := pipelinesv1.Pipeline{
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
	}

	return TestContext{
		Pipeline:                  pipeline,
		PipelineLookupKey:         types.NamespacedName{Name: pipelineName, Namespace: PipelineNamespace},
		CreationWorkflowLookupKey: types.NamespacedName{Name: "create-pipeline-" + pipelineName, Namespace: PipelineNamespace},
		UpdateWorkflowLookupKey:   types.NamespacedName{Name: "update-pipeline-" + pipelineName, Namespace: PipelineNamespace},
		DeletionWorkflowLookupKey: types.NamespacedName{Name: "delete-pipeline-" + pipelineName, Namespace: PipelineNamespace},
		Version:                   pipelinesv1.ComputeVersion(pipeline.Spec),
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

func (testCtx TestContext) pipelineExists() error {
	pipeline := &pipelinesv1.Pipeline{}
	return k8sClient.Get(ctx, testCtx.PipelineLookupKey, pipeline)
}

func (testCtx TestContext) workflowInputToMatch(name types.NamespacedName, matcher func(Gomega, map[string]string)) func(Gomega) {

	var mapParams = func(params []argo.Parameter) map[string]string {
		m := make(map[string]string)
		for i := range params {
			m[params[i].Name] = string(*params[i].Value)
		}

		return m
	}

	return func(g Gomega) {
		workflow := &argo.Workflow{}
		g.Expect(k8sClient.Get(ctx, name, workflow)).To(Succeed())

		worklfowInputParameters := mapParams(workflow.Spec.Arguments.Parameters)
		matcher(g, worklfowInputParameters)
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

func (testCtx TestContext) pipelineCreated() {
	testCtx.pipelineCreatedWithStatus(pipelinesv1.PipelineStatus{
		Id:                   PipelineId,
		SynchronizationState: pipelinesv1.Succeeded,
		Version:              testCtx.Version,
	})
}

func (testCtx TestContext) deletePipeline() error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := k8sClient.Get(ctx, testCtx.PipelineLookupKey, pipeline); err != nil {
		return err
	}

	return k8sClient.Delete(ctx, pipeline)
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
		testCtx := NewTestContext()

		expectedConfig := map[interface{}]interface{}{
			"image":         "image:v1",
			"tfxComponents": "pipeline.create_components",
			"env": map[interface{}]interface{}{
				"a": "aVal",
				"b": "bVal",
			},
		}

		It("updates the SynchronizationStatus, Id and Version", func() {
			Expect(k8sClient.Create(ctx, &testCtx.Pipeline)).To(Succeed())

			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
			})).Should(Succeed())

			Eventually(testCtx.workflowInputToMatch(testCtx.CreationWorkflowLookupKey, func(g Gomega, params map[string]string) {
				actualConfig := make(map[interface{}]interface{})
				yaml.Unmarshal([]byte(params["config"]), actualConfig)
				g.Expect(actualConfig).To(Equal(expectedConfig))
			})).Should(Succeed())

			Expect(testCtx.updateWorkflow(testCtx.CreationWorkflowLookupKey, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setWorkflowOutput(workflow, PipelineId)
			})).To(Succeed())

			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
				g.Expect(pipeline.Status.Id).To(Equal(PipelineId))
				g.Expect(pipeline.Status.Version).To(Equal(testCtx.Version))
			})).Should(Succeed())
		})
	})

	When("Creation of a pipeline fails", func() {
		testCtx := NewTestContext()

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
		testCtx := NewTestContext()

		It("keeps the status unchanged", func() {

			fmt.Println(testCtx.Pipeline.Spec)
			testCtx.pipelineCreatedWithStatus(pipelinesv1.PipelineStatus{
				Id:                   PipelineId,
				SynchronizationState: pipelinesv1.Succeeded,
				Version:              testCtx.Version,
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
		testCtx := NewTestContext()

		expectedConfig := map[interface{}]interface{}{
			"image":         "image:v1",
			"tfxComponents": "pipeline.create_components",
			"env": map[interface{}]interface{}{
				"a": "aVal",
				"b": "bVal",
				"c": "cVal",
			},
		}

		It("updates the SynchronizationStatus to Succeeded", func() {

			testCtx.pipelineCreated()

			modifiedSpec := testCtx.Pipeline.Spec
			modifiedSpec.Env["c"] = "cVal"
			modifiedVersion := pipelinesv1.ComputeVersion(modifiedSpec)

			Expect(testCtx.updatePipeline(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = modifiedSpec
			})).To(Succeed())

			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
			})).Should(Succeed())

			Eventually(testCtx.workflowInputToMatch(testCtx.UpdateWorkflowLookupKey, func(g Gomega, params map[string]string) {
				actualConfig := make(map[interface{}]interface{})
				yaml.Unmarshal([]byte(params["config"]), actualConfig)
				g.Expect(actualConfig).To(Equal(expectedConfig))
				g.Expect(params["pipeline-id"]).To(Equal(PipelineId))
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
		testCtx := NewTestContext()

		It("updates the SynchronizationStatus to Succeeded", func() {

			testCtx.pipelineCreated()

			modifiedSpec := testCtx.Pipeline.Spec
			modifiedSpec.Env["c"] = "cVal"
			modifiedVersion := pipelinesv1.ComputeVersion(modifiedSpec)

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

	When("Deletion of a pipeline succeeds", func() {
		testCtx := NewTestContext()

		It("Releases the pipeline resource", func() {
			testCtx.pipelineCreated()

			Expect(testCtx.deletePipeline()).To(Succeed())

			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Deleting))
			})).Should(Succeed())

			Eventually(testCtx.workflowInputToMatch(testCtx.DeletionWorkflowLookupKey, func(g Gomega, params map[string]string) {
				g.Expect(params["pipeline-id"]).To(Equal(PipelineId))
			})).Should(Succeed())

			Expect(testCtx.updateWorkflow(testCtx.DeletionWorkflowLookupKey, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
			})).To(Succeed())

			Eventually(testCtx.pipelineExists).Should(Not(Succeed()))
		})
	})

	When("Deletion of a pipeline fails", func() {
		testCtx := NewTestContext()

		It("Releases the pipeline resource", func() {
			testCtx.pipelineCreated()

			Expect(testCtx.deletePipeline()).To(Succeed())

			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Deleting))
			})).Should(Succeed())

			Eventually(testCtx.workflowInputToMatch(testCtx.DeletionWorkflowLookupKey, func(g Gomega, params map[string]string) {
				g.Expect(params["pipeline-id"]).To(Equal(PipelineId))
			})).Should(Succeed())

			Expect(testCtx.updateWorkflow(testCtx.DeletionWorkflowLookupKey, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowFailed
			})).To(Succeed())

			Eventually(testCtx.pipelineExists).Should(Not(Succeed()))
		})
	})
})
