//go:build decoupled || integration
// +build decoupled integration

package pipelines

import (
	"context"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	//+kubebuilder:scaffold:imports
)

type PipelineTestContext struct {
	TestContext
	Pipeline        *pipelinesv1.Pipeline
	PipelineVersion string
}

func NewPipelineTestContext(pipeline *pipelinesv1.Pipeline, k8sClient client.Client, ctx context.Context) PipelineTestContext {
	return PipelineTestContext{
		TestContext: TestContext{
			K8sClient:      k8sClient,
			ctx:            ctx,
			LookupKey:      pipeline.NamespacedName(),
			LookupLabel:    PipelineWorkflowConstants.PipelineNameLabelKey,
			operationLabel: PipelineWorkflowConstants.OperationLabelKey,
		},
		Pipeline:        pipeline,
		PipelineVersion: pipeline.Spec.ComputeVersion(),
	}
}

func (testCtx PipelineTestContext) PipelineToMatch(matcher func(Gomega, *pipelinesv1.Pipeline)) func(Gomega) {
	return func(g Gomega) {
		pipeline := &pipelinesv1.Pipeline{}
		Expect(testCtx.K8sClient.Get(testCtx.ctx, testCtx.LookupKey, pipeline)).To(Succeed())
		matcher(g, pipeline)
	}
}

func (testCtx PipelineTestContext) PipelineExists() error {
	pipeline := &pipelinesv1.Pipeline{}
	err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.LookupKey, pipeline)

	return err
}

func (testCtx PipelineTestContext) UpdatePipeline(updateFunc func(*pipelinesv1.Pipeline)) error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.LookupKey, pipeline); err != nil {
		return err
	}

	updateFunc(pipeline)

	return testCtx.K8sClient.Update(testCtx.ctx, pipeline)
}

func (testCtx PipelineTestContext) UpdatePipelineStatus(updateFunc func(*pipelinesv1.Pipeline)) error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.LookupKey, pipeline); err != nil {
		return err
	}

	updateFunc(pipeline)

	return testCtx.K8sClient.Status().Update(testCtx.ctx, pipeline)
}

func (testCtx PipelineTestContext) DeletePipeline() error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.LookupKey, pipeline); err != nil {
		return err
	}

	return testCtx.K8sClient.Delete(testCtx.ctx, pipeline)
}

func (testCtx PipelineTestContext) PipelineCreatedWithStatus(status pipelinesv1.Status) {
	Expect(testCtx.K8sClient.Create(testCtx.ctx, testCtx.Pipeline)).To(Succeed())

	Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
		g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
		g.Expect(testCtx.UpdatePipelineStatus(func(pipeline *pipelinesv1.Pipeline) {
			pipeline.Status = status
		})).To(Succeed())
	})).Should(Succeed())
}
