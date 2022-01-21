//go:build decoupled || integration
// +build decoupled integration

package pipelines

import (
	"context"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExperimentTestContext struct {
	TestContext
	Experiment *pipelinesv1.Experiment
}

func NewExperimentTestContext(experiment *pipelinesv1.Experiment, k8sClient client.Client, ctx context.Context) ExperimentTestContext {
	return ExperimentTestContext{
		TestContext: TestContext{
			K8sClient:      k8sClient,
			ctx:            ctx,
			LookupKey:      experiment.NamespacedName(),
			LookupLabel:    ExperimentWorkflowConstants.ExperimentNameLabelKey,
			operationLabel: ExperimentWorkflowConstants.OperationLabelKey,
		},
		Experiment: experiment,
	}
}

func (testCtx ExperimentTestContext) ExperimentToMatch(matcher func(Gomega, *pipelinesv1.Experiment)) func(Gomega) {
	return func(g Gomega) {
		rc := &pipelinesv1.Experiment{}
		Expect(testCtx.K8sClient.Get(testCtx.ctx, testCtx.LookupKey, rc)).To(Succeed())
		matcher(g, rc)
	}
}

func (testCtx ExperimentTestContext) ExperimentExists() error {
	rc := &pipelinesv1.Experiment{}
	err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.LookupKey, rc)

	return err
}

func (testCtx ExperimentTestContext) UpdateExperiment(updateFunc func(*pipelinesv1.Experiment)) error {
	rc := &pipelinesv1.Experiment{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.LookupKey, rc); err != nil {
		return err
	}

	updateFunc(rc)

	return testCtx.K8sClient.Update(testCtx.ctx, rc)
}

func (testCtx ExperimentTestContext) DeleteExperiment() error {
	rc := &pipelinesv1.Experiment{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.LookupKey, rc); err != nil {
		return err
	}

	return testCtx.K8sClient.Delete(testCtx.ctx, rc)
}
