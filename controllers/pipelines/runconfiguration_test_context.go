//go:build decoupled || integration
// +build decoupled integration

package pipelines

import (
	"context"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	JobId        = "12345"
	NewJobId     = "abcde"
	ExperimentId = "67890"
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
