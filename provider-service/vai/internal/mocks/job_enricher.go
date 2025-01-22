package mocks

import (
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/stretchr/testify/mock"
)

type MockJobEnricher struct{ mock.Mock }

func (m *MockJobEnricher) Enrich(
	job *aiplatformpb.PipelineJob,
	raw map[string]any,
) (*aiplatformpb.PipelineJob, error) {
	args := m.Called(job, raw)
	var pipelineJob *aiplatformpb.PipelineJob
	if arg0 := args.Get(0); arg0 != nil {
		pipelineJob = arg0.(*aiplatformpb.PipelineJob)
	}
	return pipelineJob, args.Error(1)
}
