//go:build decoupled || integration
// +build decoupled integration

package pipelines

import (
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	//+kubebuilder:scaffold:imports
)

type PipelineCreate struct {
}

func (pt PipelineCreate) new() *pipelinesv1.Pipeline {
	return &pipelinesv1.Pipeline{}
}

func NewPipelineTestContext(pipeline *pipelinesv1.Pipeline) TestContext[*pipelinesv1.Pipeline] {
	return TestContext[*pipelinesv1.Pipeline]{
		K8sClient:      k8sClient,
		ctx:            ctx,
		OwnerKind:      PipelineWorkflowConstants.PipelineKind,
		NamespacedName: pipeline.NamespacedName(),
		Create:         PipelineCreate{},
		Resource:       pipeline,
	}
}
