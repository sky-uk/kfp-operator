//go:build decoupled || integration
// +build decoupled integration

package pipelines

import (
	"context"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
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
			OwnerName:      pipeline.Name,
			OwnerKind: "pipeline",
		},
		Pipeline:        pipeline,
		PipelineVersion: pipeline.Spec.ComputeVersion(),
	}
}

func (testCtx PipelineTestContext) PipelineToMatch(matcher func(Gomega, *pipelinesv1.Pipeline)) func(Gomega) {
	return func(g Gomega) {
		pipeline := &pipelinesv1.Pipeline{}
		Expect(testCtx.K8sClient.Get(testCtx.ctx, testCtx.Pipeline.NamespacedName(), pipeline)).To(Succeed())
		matcher(g, pipeline)
	}
}

func (testCtx PipelineTestContext) PipelineExists() error {
	pipeline := &pipelinesv1.Pipeline{}
	err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.Pipeline.NamespacedName(), pipeline)

	return err
}

func (testCtx PipelineTestContext) UpdatePipeline(updateFunc func(*pipelinesv1.Pipeline)) error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.Pipeline.NamespacedName(), pipeline); err != nil {
		return err
	}

	updateFunc(pipeline)

	return testCtx.K8sClient.Update(testCtx.ctx, pipeline)
}

func (testCtx PipelineTestContext) UpdatePipelineStatus(updateFunc func(*pipelinesv1.Pipeline)) error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.Pipeline.NamespacedName(), pipeline); err != nil {
		return err
	}

	updateFunc(pipeline)

	return testCtx.K8sClient.Status().Update(testCtx.ctx, pipeline)
}

func (testCtx PipelineTestContext) DeletePipeline() error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.Pipeline.NamespacedName(), pipeline); err != nil {
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

func (testCtx PipelineTestContext) EmittedEventsToMatch(matcher func(Gomega, []v1.Event)) func(Gomega) {
	return func(g Gomega) {
		eventList := &v1.EventList{}
		Expect(testCtx.K8sClient.List(testCtx.ctx, eventList, client.MatchingFields{"involvedObject.name": testCtx.Pipeline.Name})).To(Succeed())

		matcher(g, eventList.Items)
	}
}
