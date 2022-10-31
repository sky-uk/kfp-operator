//go:build decoupled || integration
// +build decoupled integration

package pipelines

import (
	"context"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	//+kubebuilder:scaffold:imports
)

type PipelineTestContext struct {
	TestContext
	Pipeline *pipelinesv1.Pipeline
}

func NewPipelineTestContext(pipeline *pipelinesv1.Pipeline, k8sClient client.Client, ctx context.Context) PipelineTestContext {
	return PipelineTestContext{
		TestContext: TestContext{
			K8sClient: k8sClient,
			ctx:       ctx,
			Resource:  pipeline,
		},
		Pipeline: pipeline,
	}
}

func (testCtx PipelineTestContext) PipelineToMatch(matcher func(Gomega, *pipelinesv1.Pipeline)) func(Gomega) {
	return func(g Gomega) {
		pipeline := &pipelinesv1.Pipeline{}
		Expect(testCtx.K8sClient.Get(testCtx.ctx, testCtx.Resource.GetNamespacedName(), pipeline)).To(Succeed())
		matcher(g, pipeline)
	}
}

func (testCtx PipelineTestContext) PipelineExists() error {
	pipeline := &pipelinesv1.Pipeline{}
	err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.Resource.GetNamespacedName(), pipeline)

	return err
}

func (testCtx PipelineTestContext) UpdatePipeline(updateFunc func(*pipelinesv1.Pipeline)) error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.Resource.GetNamespacedName(), pipeline); err != nil {
		return err
	}

	updateFunc(pipeline)

	return testCtx.K8sClient.Update(testCtx.ctx, pipeline)
}

func (testCtx PipelineTestContext) UpdatePipelineStatus(updateFunc func(*pipelinesv1.Pipeline)) error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.Resource.GetNamespacedName(), pipeline); err != nil {
		return err
	}

	updateFunc(pipeline)

	return testCtx.K8sClient.Status().Update(testCtx.ctx, pipeline)
}

func (testCtx PipelineTestContext) DeletePipeline() error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.Resource.GetNamespacedName(), pipeline); err != nil {
		return err
	}

	return testCtx.K8sClient.Delete(testCtx.ctx, pipeline)
}

func (testCtx PipelineTestContext) StablePipelineCreated() {
	testCtx.PipelineCreatedWithStatus(pipelinesv1.Status{
		Version:              testCtx.Pipeline.ComputeVersion(),
		ProviderId:           apis.RandomString(),
		SynchronizationState: apis.Succeeded,
	})
}

func (testCtx PipelineTestContext) PipelineCreatedWithStatus(status pipelinesv1.Status) {
	Expect(testCtx.K8sClient.Create(testCtx.ctx, testCtx.Pipeline)).To(Succeed())

	Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
		g.Expect(pipeline.Status.SynchronizationState).To(Equal(apis.Creating))
		g.Expect(testCtx.UpdatePipelineStatus(func(pipeline *pipelinesv1.Pipeline) {
			pipeline.Status = status
		})).To(Succeed())
	})).Should(Succeed())
}

func (testCtx PipelineTestContext) StablePipelineUpdated(pipeline pipelinesv1.Pipeline) {
	testCtx.PipelineUpdatedWithStatus(pipeline.Spec, pipelinesv1.Status{
		Version:              pipeline.ComputeVersion(),
		ProviderId:           apis.RandomString(),
		SynchronizationState: apis.Succeeded,
	})
}

func (testCtx PipelineTestContext) PipelineUpdatedWithStatus(spec pipelinesv1.PipelineSpec, status pipelinesv1.Status) {
	Expect(testCtx.UpdatePipeline(func(pipeline *pipelinesv1.Pipeline) {
		pipeline.Spec = spec
	})).To(Succeed())

	Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
		g.Expect(pipeline.Status.SynchronizationState).To(Equal(apis.Updating))
		g.Expect(testCtx.UpdatePipelineStatus(func(pipeline *pipelinesv1.Pipeline) {
			pipeline.Status = status
		})).To(Succeed())
	})).Should(Succeed())
}

func (testCtx PipelineTestContext) EmittedEventsToMatch(matcher func(Gomega, []v1.Event)) func(Gomega) {
	return func(g Gomega) {
		eventList := &v1.EventList{}
		Expect(testCtx.K8sClient.List(testCtx.ctx, eventList, client.MatchingFields{"involvedObject.name": testCtx.Resource.GetName()})).To(Succeed())

		matcher(g, eventList.Items)
	}
}
