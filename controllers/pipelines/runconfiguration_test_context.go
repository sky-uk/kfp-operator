//go:build decoupled || integration
// +build decoupled integration

package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RunConfigurationTestContext struct {
	TestContext
	RunConfiguration *pipelinesv1.RunConfiguration
}

func NewRunConfigurationTestContext(runConfiguration *pipelinesv1.RunConfiguration, k8sClient client.Client, ctx context.Context) RunConfigurationTestContext {
	return RunConfigurationTestContext{
		TestContext: TestContext{
			K8sClient:      k8sClient,
			ctx:            ctx,
			LookupKey:      runConfiguration.NamespacedName(),
			LookupLabel:    RunConfigurationWorkflowConstants.RunConfigurationNameLabelKey,
			operationLabel: RunConfigurationWorkflowConstants.OperationLabelKey,
		},
		RunConfiguration: runConfiguration,
	}
}

func (testCtx RunConfigurationTestContext) RunConfigurationToMatch(matcher func(Gomega, *pipelinesv1.RunConfiguration)) func(Gomega) {
	return func(g Gomega) {
		rc := &pipelinesv1.RunConfiguration{}
		Expect(testCtx.K8sClient.Get(testCtx.ctx, testCtx.LookupKey, rc)).To(Succeed())
		matcher(g, rc)
	}
}

func (testCtx RunConfigurationTestContext) RunConfigurationExists() error {
	rc := &pipelinesv1.RunConfiguration{}
	err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.LookupKey, rc)

	return err
}

func (testCtx RunConfigurationTestContext) UpdateRunConfiguration(updateFunc func(*pipelinesv1.RunConfiguration)) error {
	rc := &pipelinesv1.RunConfiguration{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.LookupKey, rc); err != nil {
		return err
	}

	updateFunc(rc)

	return testCtx.K8sClient.Update(testCtx.ctx, rc)
}

func (testCtx RunConfigurationTestContext) DeleteRunConfiguration() error {
	rc := &pipelinesv1.RunConfiguration{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.LookupKey, rc); err != nil {
		return err
	}

	return testCtx.K8sClient.Delete(testCtx.ctx, rc)
}

func (testCtx RunConfigurationTestContext) EmittedEventsToMatch(matcher func(Gomega, []v1.Event)) func(Gomega) {
	return func(g Gomega) {
		eventList := &v1.EventList{}
		Expect(testCtx.K8sClient.List(testCtx.ctx, eventList, client.MatchingFields{"involvedObject.name": testCtx.RunConfiguration.Name})).To(Succeed())

		matcher(g, eventList.Items)
	}
}

func (testCtx RunConfigurationTestContext) WorkflowSucceeded(operation string) {
	Eventually(testCtx.WorkflowToBeUpdated(operation, func(workflow *argo.Workflow) {
		workflow.Status.Phase = argo.WorkflowSucceeded
		setWorkflowOutputs(
			workflow,
			[]argo.Parameter{
				{
					Name:  RunConfigurationWorkflowConstants.RunConfigurationIdParameterName,
					Value: argo.AnyStringPtr(RandomString()),
				},
			},
		)
	})).Should(Succeed())
}

func (testCtx RunConfigurationTestContext) UpdateRunConfigurationStatus(updateFunc func(configuration *pipelinesv1.RunConfiguration)) error {
	runConfiguration := &pipelinesv1.RunConfiguration{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.LookupKey, runConfiguration); err != nil {
		return err
	}

	updateFunc(runConfiguration)

	return testCtx.K8sClient.Status().Update(testCtx.ctx, runConfiguration)
}

func (testCtx RunConfigurationTestContext) RunConfigurationCreatedWithStatus(status pipelinesv1.RunConfigurationStatus) {
	Expect(testCtx.K8sClient.Create(testCtx.ctx, testCtx.RunConfiguration)).To(Succeed())

	Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, pipeline *pipelinesv1.RunConfiguration) {
		g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
		g.Expect(testCtx.UpdateRunConfigurationStatus(func(pipeline *pipelinesv1.RunConfiguration) {
			pipeline.Status = status
		})).To(Succeed())
	})).Should(Succeed())
}
