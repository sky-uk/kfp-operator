//go:build decoupled || integration
// +build decoupled integration

package pipelines

import (
	"context"
	"errors"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	//+kubebuilder:scaffold:imports
)

type TestContext struct {
	K8sClient         client.Client
	ctx               context.Context
	LookupKey     	  types.NamespacedName
	LookupLabel       string
}

func (testCtx TestContext) WorkflowInputToMatch(operation string, matcher func(Gomega, map[string]string)) func(Gomega) {

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

func (testCtx TestContext) WorkflowToMatch(operation string, matcher func(Gomega, *argo.Workflow)) func(Gomega) {

	return func(g Gomega) {
		workflow, err := testCtx.fetchWorkflow(operation)

		Expect(err).NotTo(HaveOccurred())

		matcher(g, workflow)
	}
}

func (testCtx TestContext) UpdateWorkflow(operation string, updateFunc func(*argo.Workflow)) error {
	workflow, err := testCtx.fetchWorkflow(operation)
	if err != nil {
		return err
	}

	updateFunc(workflow)
	return testCtx.K8sClient.Update(testCtx.ctx, workflow)
}

func (testCtx TestContext) fetchWorkflow(operation string) (*argo.Workflow, error) {
	workflowList := &argo.WorkflowList{}

	if err := testCtx.K8sClient.List(testCtx.ctx, workflowList, client.MatchingLabels{PipelineWorkflowConstants.OperationLabelKey: operation, testCtx.LookupLabel: testCtx.LookupKey.Name}); err != nil {
		return nil, err
	}

	if len(workflowList.Items) != 1 {
		return nil, errors.New("not exactly one workflow")
	}

	return &workflowList.Items[0], nil
}
