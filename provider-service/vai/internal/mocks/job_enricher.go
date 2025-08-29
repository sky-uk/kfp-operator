//go:build unit

package mocks

import (
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/stretchr/testify/mock"
)

type MockJobEnricher struct{ mock.Mock }

func (m *MockJobEnricher) Enrich(
	job *aiplatformpb.PipelineJob,
	compiledPipeline resource.CompiledPipeline,
) (*aiplatformpb.PipelineJob, error) {
	args := m.Called(job, compiledPipeline)
	var pipelineJob *aiplatformpb.PipelineJob
	if arg0 := args.Get(0); arg0 != nil {
		pipelineJob = arg0.(*aiplatformpb.PipelineJob)
	}
	return pipelineJob, args.Error(1)
}
