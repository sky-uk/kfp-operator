//go:build decoupled || integration

package testutil

import (
	"errors"
	"fmt"

	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//+kubebuilder:scaffold:imports
)

type DeploymentTestHelper[R pipelinesv1.Resource] struct {
	Resource R
}

func (dth DeploymentTestHelper[R]) UpdateDeployment(updateFunc func(*appsv1.Deployment)) func(g Gomega) {
	return func(g Gomega) {
		g.Expect(dth.updateDeployment(updateFunc)).To(Succeed())
	}
}

func (dth DeploymentTestHelper[R]) updateDeployment(updateFunc func(*appsv1.Deployment)) error {
	deployment, err := dth.fetchDeployment()
	if err != nil {
		return err
	}

	updateFunc(deployment)
	return K8sClient.Update(Ctx, deployment)

}

func (dth DeploymentTestHelper[R]) fetchDeployment() (*appsv1.Deployment, error) {
	deploymentList := &appsv1.DeploymentList{}
	if err := K8sClient.List(
		Ctx,
		deploymentList,
		// TODO: should probably use the OwnerNameLabel const here (but having issues with it)
		client.MatchingLabels(map[string]string{"owner-name": fmt.Sprintf("provider-%s", dth.Resource.GetName())}),
	); err != nil {
		return nil, err
	}

	for _, deployment := range deploymentList.Items {
		if metav1.IsControlledBy(&deployment, dth.Resource) {
			return &deployment, nil
		}
	}

	return nil, errors.New("no deployment found")
}

// func (testCtx WorkflowTestHelper[R]) WorkflowByNameToMatch(
// 	namespacedName types.NamespacedName,
// 	matcher func(Gomega, *argo.Workflow),
// ) func(Gomega) {
//
// 	return func(g Gomega) {
// 		workflow := &argo.Workflow{}
// 		Expect(K8sClient.Get(Ctx, namespacedName, workflow)).To(Succeed())
//
// 		matcher(g, workflow)
// 	}
// }
//
// func (testCtx WorkflowTestHelper[R]) UpdateWorkflow(updateFunc func(*argo.Workflow)) error {
// 	workflow, err := testCtx.fetchWorkflow()
//
// 	if err != nil {
// 		return err
// 	}
//
// 	updateFunc(workflow)
// 	return K8sClient.Update(Ctx, workflow)
// }
//
//
// func (testCtx WorkflowTestHelper[R]) FetchWorkflow() func() error {
// 	return func() error {
// 		_, err := testCtx.fetchWorkflow()
// 		return err
// 	}
// }
//
// func (testCtx WorkflowTestHelper[R]) fetchWorkflow() (*argo.Workflow, error) {
// 	workflowList := &argo.WorkflowList{}
//
// 	if err := K8sClient.List(
// 		Ctx,
// 		workflowList,
// 		client.MatchingLabels(workflowconstants.CommonWorkflowLabels(testCtx.Resource)),
// 	); err != nil {
// 		return nil, err
// 	}
//
// 	for _, workflow := range workflowList.Items {
// 		if workflow.Status.Phase != argo.WorkflowSucceeded {
// 			return &workflow, nil
// 		}
// 	}
//
// 	return nil, errors.New("no workflow found")
// }
