//go:build decoupled || integration
// +build decoupled integration

package pipelines

import (
	"context"
	"errors"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//+kubebuilder:scaffold:imports
)

var (
	k8sClient client.Client
	ctx       context.Context
)

type WorkflowTestHelper[R pipelinesv1.Resource] struct {
	Resource R
}

func (testCtx WorkflowTestHelper[R]) WorkflowInputToMatch(operation string, matcher func(Gomega, map[string]string)) func(Gomega) {

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

func (testCtx WorkflowTestHelper[R]) WorkflowByNameToMatch(namespacedName types.NamespacedName, matcher func(Gomega, *argo.Workflow)) func(Gomega) {

	return func(g Gomega) {
		workflow := &argo.Workflow{}
		Expect(k8sClient.Get(ctx, namespacedName, workflow)).To(Succeed())

		matcher(g, workflow)
	}
}

func (testCtx WorkflowTestHelper[R]) WorkflowByOperationToMatch(operation string, matcher func(Gomega, *argo.Workflow)) func(Gomega) {

	return func(g Gomega) {
		workflow, err := testCtx.fetchWorkflow(operation)

		Expect(err).NotTo(HaveOccurred())

		matcher(g, workflow)
	}
}

func (testCtx WorkflowTestHelper[R]) UpdateWorkflow(operation string, updateFunc func(*argo.Workflow)) error {
	workflow, err := testCtx.fetchWorkflow(operation)

	if err != nil {
		return err
	}

	updateFunc(workflow)
	return k8sClient.Update(ctx, workflow)
}

func (testCtx WorkflowTestHelper[R]) WorkflowToBeUpdated(operation string, updateFunc func(*argo.Workflow)) func(g Gomega) {
	return func(g Gomega) {
		g.Expect(testCtx.UpdateWorkflow(operation, updateFunc)).To(Succeed())
	}
}

func (testCtx WorkflowTestHelper[R]) FetchWorkflow(operation string) func() error {
	return func() error {
		_, err := testCtx.fetchWorkflow(operation)
		return err
	}
}

func (testCtx WorkflowTestHelper[R]) fetchWorkflow(operation string) (*argo.Workflow, error) {
	workflowList := &argo.WorkflowList{}

	if err := k8sClient.List(ctx, workflowList, client.MatchingLabels(CommonWorkflowLabels(testCtx.Resource, operation))); err != nil {
		return nil, err
	}

	numberOfWorkflows := len(workflowList.Items)
	if numberOfWorkflows != 1 {
		return nil, errors.New(fmt.Sprintf("Want exactly 1 workflow. Have %d", numberOfWorkflows))
	}

	return &workflowList.Items[0], nil
}
