//go:build decoupled || integration
// +build decoupled integration

package pipelines

import (
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

type ExperimentCreate struct {
}

func (pt ExperimentCreate) new() *pipelinesv1.Experiment {
	return &pipelinesv1.Experiment{}
}

func NewExperimentTestContext(experiment *pipelinesv1.Experiment) TestContext[*pipelinesv1.Experiment] {
	return TestContext[*pipelinesv1.Experiment]{
		K8sClient:      k8sClient,
		ctx:            ctx,
		OwnerKind:      ExperimentWorkflowConstants.ExperimentKind,
		NamespacedName: experiment.NamespacedName(),
		Create:         ExperimentCreate{},
		Resource:       experiment,
	}
}
