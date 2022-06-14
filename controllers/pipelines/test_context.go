//go:build decoupled || integration
// +build decoupled integration

package pipelines

import (
	"context"
	"errors"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//+kubebuilder:scaffold:imports
)

var (
	k8sClient client.Client
	ctx       context.Context
)

type Create[T Resource] interface {
	new() T
}

type TestContext[T Resource] struct {
	K8sClient      client.Client
	ctx            context.Context
	OwnerKind      string
	NamespacedName types.NamespacedName
	Resource       T
	Create         Create[T]
}

func (testCtx TestContext[T]) CreateResource() error {
	return k8sClient.Create(ctx, testCtx.Resource)
}

func (testCtx TestContext[T]) ResourceToMatch(matcher func(Gomega, T)) func(Gomega) {
	return func(g Gomega) {
		t := testCtx.Create.new()
		Expect(testCtx.K8sClient.Get(testCtx.ctx, testCtx.NamespacedName, t)).To(Succeed())
		matcher(g, t)
	}
}

func (testCtx TestContext[T]) ResourceExists() error {
	t := testCtx.Create.new()
	err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.NamespacedName, t)
	return err
}

func (testCtx TestContext[T]) UpdateResource(updateFunc func(T)) error {
	t := testCtx.Create.new()

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.NamespacedName, t); err != nil {
		return err
	}

	updateFunc(t)

	return testCtx.K8sClient.Update(testCtx.ctx, t)
}

func (testCtx TestContext[T]) UpdateResourceStatus(updateFunc func(T)) error {
	t := testCtx.Create.new()

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.NamespacedName, t); err != nil {
		return err
	}

	updateFunc(t)

	return testCtx.K8sClient.Status().Update(testCtx.ctx, t)
}

func (testCtx TestContext[T]) DeleteResource() error {
	t := testCtx.Create.new()

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.NamespacedName, t); err != nil {
		return err
	}

	return testCtx.K8sClient.Delete(testCtx.ctx, t)
}

func (testCtx TestContext[T]) ResourceCreatedWithStatus(status pipelinesv1.Status) {
	Expect(testCtx.K8sClient.Create(testCtx.ctx, testCtx.Resource)).To(Succeed())

	Eventually(testCtx.ResourceToMatch(func(g Gomega, t T) {
		g.Expect(t.GetStatus().SynchronizationState).To(Equal(pipelinesv1.Creating))
		g.Expect(testCtx.UpdateResourceStatus(func(t T) {
			t.SetStatus(status)
		})).To(Succeed())
	})).Should(Succeed())
}

func (testCtx TestContext[T]) EmittedEventsToMatch(matcher func(Gomega, []v1.Event)) func(Gomega) {
	return func(g Gomega) {
		eventList := &v1.EventList{}
		Expect(testCtx.K8sClient.List(testCtx.ctx, eventList, client.MatchingFields{"involvedObject.name": testCtx.NamespacedName.Name})).To(Succeed())

		matcher(g, eventList.Items)
	}
}

func (testCtx TestContext[T]) WorkflowInputToMatch(operation string, matcher func(Gomega, map[string]string)) func(Gomega) {

	var mapParams = func(params []argo.Parameter) map[string]string {
		m := make(map[string]string, len(params))
		for i := range params {
			m[params[i].Name] = string(*params[i].Value)
		}

		return m
	}

	return func(g Gomega) {
		workflow, err := testCtx.fetchWorkflow(operation)

		Expect(err).NotTo(HaveOccurred())

		worklfowInputParameters := mapParams(workflow.Spec.Arguments.Parameters)
		matcher(g, worklfowInputParameters)
	}
}

func (testCtx TestContext[T]) WorkflowByNameToMatch(namespacedName types.NamespacedName, matcher func(Gomega, *argo.Workflow)) func(Gomega) {

	return func(g Gomega) {
		workflow := &argo.Workflow{}
		Expect(testCtx.K8sClient.Get(testCtx.ctx, namespacedName, workflow)).To(Succeed())

		matcher(g, workflow)
	}
}

func (testCtx TestContext[T]) WorkflowByOperationToMatch(operation string, matcher func(Gomega, *argo.Workflow)) func(Gomega) {

	return func(g Gomega) {
		workflow, err := testCtx.fetchWorkflow(operation)

		Expect(err).NotTo(HaveOccurred())

		matcher(g, workflow)
	}
}

func (testCtx TestContext[T]) UpdateWorkflow(operation string, updateFunc func(*argo.Workflow)) error {
	workflow, err := testCtx.fetchWorkflow(operation)

	if err != nil {
		return err
	}

	updateFunc(workflow)
	return testCtx.K8sClient.Update(testCtx.ctx, workflow)
}

func (testCtx TestContext[T]) WorkflowToBeUpdated(operation string, updateFunc func(*argo.Workflow)) func(g Gomega) {
	return func(g Gomega) {
		g.Expect(testCtx.UpdateWorkflow(operation, updateFunc)).To(Succeed())
	}
}

func (testCtx TestContext[T]) FetchWorkflow(operation string) func() error {
	return func() error {
		_, err := testCtx.fetchWorkflow(operation)
		return err
	}
}

func (testCtx TestContext[T]) fetchWorkflow(operation string) (*argo.Workflow, error) {
	workflowList := &argo.WorkflowList{}

	if err := testCtx.K8sClient.List(testCtx.ctx, workflowList, client.MatchingLabels(CommonWorkflowLabels(testCtx.NamespacedName, operation, testCtx.OwnerKind))); err != nil {
		return nil, err
	}

	numberOfWorkflows := len(workflowList.Items)
	if numberOfWorkflows != 1 {
		return nil, errors.New(fmt.Sprintf("Want exactly 1 workflow. Have %d", numberOfWorkflows))
	}

	return &workflowList.Items[0], nil
}
