//go:build unit
// +build unit

package main
//
//import (
//	"context"
//	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
//	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
//	"testing"
//
//	. "github.com/onsi/ginkgo"
//	. "github.com/onsi/gomega"
//)
//
//func TestUnitSuite(t *testing.T) {
//	RegisterFailHandler(Fail)
//	RunSpecs(t, "Model Update Event Source Unit Suite")
//}
//
//func workflowInPhase(phase argo.WorkflowPhase) *unstructured.Unstructured {
//	workflow := unstructured.Unstructured{}
//	workflow.SetName("test-workflow")
//	workflow.SetLabels(map[string]string{
//		"workflows.argoproj.io/phase": string(phase),
//	})
//
//	return &workflow
//}
//
//var _ = Describe("Event for workflow update", func() {
//	anEvent := &api.Event{
//		Name:    "model-update",
//		Payload: []byte("test-workflow"),
//	}
//
//	When("the workflow succeeded", func() {
//		oldObj := workflowInPhase(argo.WorkflowPending)
//		newObj := workflowInPhase(argo.WorkflowSucceeded)
//		event := eventForWorkflowUpdate(context.Background(), oldObj, newObj)
//
//		It("creates an event", func() {
//			Expect(event).To(Equal(anEvent))
//		})
//	})
//
//	When("the workflow stays in succeeded", func() {
//		oldObj := workflowInPhase(argo.WorkflowSucceeded)
//		newObj := workflowInPhase(argo.WorkflowSucceeded)
//		event := eventForWorkflowUpdate(context.Background(), oldObj, newObj)
//
//		It("creates no event", func() {
//			Expect(event).To(BeNil())
//		})
//	})
//})
