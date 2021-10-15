//go:build decoupled || integration
// +build decoupled integration

package pipelines

import (
	"context"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	JobId        = "12345"
	NewJobId     = "abcde"
	ExperimentId = "67890"
)

type RunconfigurationTestContext struct {
	TestContext
	RunConfiguration *pipelinesv1.RunConfiguration
}

func NewRunconfigurationTestContext(runConfiguration *pipelinesv1.RunConfiguration, k8sClient client.Client, ctx context.Context) RunconfigurationTestContext {
	return RunconfigurationTestContext{
		TestContext: TestContext{
			K8sClient:   k8sClient,
			ctx:         ctx,
			LookupKey:   runConfiguration.NamespacedName(),
			LookupLabel: RunConfigurationWorkflowConstants.RunConfigurationNameLabelKey,
		},
		RunConfiguration: runConfiguration,
	}
}
