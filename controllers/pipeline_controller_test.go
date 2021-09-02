package controllers

import (
	// 	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"
	// 	pipelinesv1 "github.com/sky-uk/kfp-operator/api/v1"
	// 	"gopkg.in/yaml.v2"
	// 	// +kubebuilder:scaffold:imports
)

var _ = Describe("Pipeline controller", func() {
})

// 	When("Creation of a pipeline succeeds", func() {
// 		testCtx := NewTestContext()

// 		expectedConfig := map[interface{}]interface{}{
// 			"image":         "image:v1",
// 			"tfxComponents": "pipeline.create_components",
// 			"env": map[interface{}]interface{}{
// 				"a": "aVal",
// 				"b": "bVal",
// 			},
// 		}

// 		It("updates the SynchronizationStatus, Id and Version", func() {
// 			Expect(k8sClient.Create(ctx, &testCtx.Pipeline)).To(Succeed())

// 			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
// 				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
// 			})).Should(Succeed())

// 			Eventually(testCtx.workflowInputToMatch(testCtx.CreationWorkflowLookupKey, func(g Gomega, params map[string]string) {
// 				actualConfig := make(map[interface{}]interface{})
// 				yaml.Unmarshal([]byte(params["config"]), actualConfig)
// 				g.Expect(actualConfig).To(Equal(expectedConfig))
// 			})).Should(Succeed())

// 			Expect(testCtx.updateWorkflow(testCtx.CreationWorkflowLookupKey, func(workflow *argo.Workflow) {
// 				workflow.Status.Phase = argo.WorkflowSucceeded
// 				setWorkflowOutput(workflow, PipelineIdKey, PipelineId)
// 			})).To(Succeed())

// 			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
// 				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
// 				g.Expect(pipeline.Status.Id).To(Equal(PipelineId))
// 				g.Expect(pipeline.Status.Version).To(Equal(testCtx.Version))
// 			})).Should(Succeed())
// 		})
// 	})

// 	When("Creation of a pipeline fails", func() {
// 		testCtx := NewTestContext()

// 		It("updates the SynchronizationStatus to failed", func() {
// 			Expect(k8sClient.Create(ctx, &testCtx.Pipeline)).To(Succeed())

// 			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
// 				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
// 			})).Should(Succeed())

// 			Expect(testCtx.updateWorkflow(testCtx.CreationWorkflowLookupKey, func(workflow *argo.Workflow) {
// 				workflow.Status.Phase = argo.WorkflowFailed
// 			})).To(Succeed())

// 			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
// 				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Failed))
// 			})).Should(Succeed())
// 		})
// 	})

// 	When("The resource is updated with no changes", func() {
// 		testCtx := NewTestContext()

// 		It("keeps the status unchanged", func() {
// 			testCtx.pipelineCreatedWithStatus(pipelinesv1.PipelineStatus{
// 				Id:                   PipelineId,
// 				SynchronizationState: pipelinesv1.Succeeded,
// 				Version:              testCtx.Version,
// 			})

// 			Expect(testCtx.updatePipeline(func(pipeline *pipelinesv1.Pipeline) {
// 				// No changes
// 			})).To(Succeed())

// 			Consistently(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
// 				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
// 			})).Should(Succeed())
// 		})
// 	})

// 	When("The resource is updated with changes and the update succeeds", func() {
// 		testCtx := NewTestContext()

// 		expectedConfig := map[interface{}]interface{}{
// 			"image":         "image:v1",
// 			"tfxComponents": "pipeline.create_components",
// 			"env": map[interface{}]interface{}{
// 				"a": "aVal",
// 				"b": "bVal",
// 				"c": "cVal",
// 			},
// 		}

// 		It("updates the SynchronizationStatus to Succeeded", func() {

// 			testCtx.pipelineCreated()

// 			modifiedSpec := testCtx.Pipeline.Spec
// 			modifiedSpec.Env["c"] = "cVal"
// 			modifiedVersion := pipelinesv1.ComputeVersion(modifiedSpec)

// 			Expect(testCtx.updatePipeline(func(pipeline *pipelinesv1.Pipeline) {
// 				pipeline.Spec = modifiedSpec
// 			})).To(Succeed())

// 			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
// 				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
// 			})).Should(Succeed())

// 			Eventually(testCtx.workflowInputToMatch(testCtx.UpdateWorkflowLookupKey, func(g Gomega, params map[string]string) {
// 				actualConfig := make(map[interface{}]interface{})
// 				yaml.Unmarshal([]byte(params["config"]), actualConfig)
// 				g.Expect(actualConfig).To(Equal(expectedConfig))
// 				g.Expect(params["pipeline-id"]).To(Equal(PipelineId))
// 			})).Should(Succeed())

// 			Expect(testCtx.updateWorkflow(testCtx.UpdateWorkflowLookupKey, func(workflow *argo.Workflow) {
// 				workflow.Status.Phase = argo.WorkflowSucceeded
// 			})).To(Succeed())

// 			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
// 				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
// 				g.Expect(pipeline.Status.Version).To(Equal(modifiedVersion))
// 			})).Should(Succeed())
// 		})
// 	})

// 	When("The resource is updated with changes and the update fails", func() {
// 		testCtx := NewTestContext()

// 		It("updates the SynchronizationStatus to Failed", func() {

// 			testCtx.pipelineCreated()

// 			modifiedSpec := testCtx.Pipeline.Spec
// 			modifiedSpec.Env["c"] = "cVal"
// 			modifiedVersion := pipelinesv1.ComputeVersion(modifiedSpec)

// 			Expect(testCtx.updatePipeline(func(pipeline *pipelinesv1.Pipeline) {
// 				pipeline.Spec = modifiedSpec
// 			})).To(Succeed())

// 			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
// 				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
// 			})).Should(Succeed())

// 			Expect(testCtx.updateWorkflow(testCtx.UpdateWorkflowLookupKey, func(workflow *argo.Workflow) {
// 				workflow.Status.Phase = argo.WorkflowFailed
// 			})).To(Succeed())

// 			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
// 				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Failed))
// 				g.Expect(pipeline.Status.Version).To(Equal(modifiedVersion))
// 			})).Should(Succeed())
// 		})
// 	})

// 	When("Deletion of a pipeline succeeds", func() {
// 		testCtx := NewTestContext()

// 		It("Releases the pipeline resource", func() {
// 			testCtx.pipelineCreated()

// 			Expect(testCtx.deletePipeline()).To(Succeed())

// 			Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
// 				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Deleting))
// 			})).Should(Succeed())

// 			Eventually(testCtx.workflowInputToMatch(testCtx.DeletionWorkflowLookupKey, func(g Gomega, params map[string]string) {
// 				g.Expect(params["pipeline-id"]).To(Equal(PipelineId))
// 			})).Should(Succeed())

// 			Expect(testCtx.updateWorkflow(testCtx.DeletionWorkflowLookupKey, func(workflow *argo.Workflow) {
// 				workflow.Status.Phase = argo.WorkflowSucceeded
// 			})).To(Succeed())

// 			Eventually(testCtx.pipelineExists).Should(Not(Succeed()))
// 		})
// 	})
// })
