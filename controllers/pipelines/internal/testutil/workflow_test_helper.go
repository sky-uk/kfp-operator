//go:build decoupled || integration

package testutil

import (
	"errors"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	workflowfactory "github.com/sky-uk/kfp-operator/controllers/pipelines"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//+kubebuilder:scaffold:imports
)

type WorkflowTestHelper[R pipelinesv1.Resource] struct {
	Resource R
}

func (testCtx WorkflowTestHelper[R]) WorkflowByNameToMatch(namespacedName types.NamespacedName, matcher func(Gomega, *argo.Workflow)) func(Gomega) {

	return func(g Gomega) {
		workflow := &argo.Workflow{}
		Expect(k8sClient.Get(ctx, namespacedName, workflow)).To(Succeed())

		matcher(g, workflow)
	}
}

func (testCtx WorkflowTestHelper[R]) UpdateWorkflow(updateFunc func(*argo.Workflow)) error {
	workflow, err := testCtx.fetchWorkflow()

	if err != nil {
		return err
	}

	updateFunc(workflow)
	return k8sClient.Update(ctx, workflow)
}

func (testCtx WorkflowTestHelper[R]) WorkflowToBeUpdated(updateFunc func(*argo.Workflow)) func(g Gomega) {
	return func(g Gomega) {
		g.Expect(testCtx.UpdateWorkflow(updateFunc)).To(Succeed())
	}
}

func (testCtx WorkflowTestHelper[R]) FetchWorkflow() func() error {
	return func() error {
		_, err := testCtx.fetchWorkflow()
		return err
	}
}

func (testCtx WorkflowTestHelper[R]) fetchWorkflow() (*argo.Workflow, error) {
	workflowList := &argo.WorkflowList{}

	if err := k8sClient.List(ctx, workflowList, client.MatchingLabels(workflowfactory.CommonWorkflowLabels(testCtx.Resource))); err != nil {
		return nil, err
	}

	for _, workflow := range workflowList.Items {
		if workflow.Status.Phase != argo.WorkflowSucceeded {
			return &workflow, nil
		}
	}

	return nil, errors.New("no workflow found")
}
