//go:build decoupled || integration
// +build decoupled integration

package pipelines

import (
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

type RunConfigurationCreate struct {
}

func (pt RunConfigurationCreate) new() *pipelinesv1.RunConfiguration {
	return &pipelinesv1.RunConfiguration{}
}

func NewRunConfigurationTestContext(runConfiguration *pipelinesv1.RunConfiguration) TestContext[*pipelinesv1.RunConfiguration] {
	return TestContext[*pipelinesv1.RunConfiguration]{
		K8sClient:      k8sClient,
		ctx:            ctx,
		OwnerKind:      RunConfigurationWorkflowConstants.RunConfigurationKind,
		NamespacedName: runConfiguration.NamespacedName(),
		Create:         RunConfigurationCreate{},
		Resource:       runConfiguration,
	}
}
