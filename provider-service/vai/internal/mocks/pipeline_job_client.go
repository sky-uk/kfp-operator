//go:build unit

package mocks

import (
	"context"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/mock"
)

type MockPipelineJobClient struct {
	mock.Mock
}

func (m *MockPipelineJobClient) GetPipelineJob(
	_ context.Context,
	req *aiplatformpb.GetPipelineJobRequest,
	_ ...gax.CallOption,
) (*aiplatformpb.PipelineJob, error) {
	args := m.Called(req)
	var pipelineJob *aiplatformpb.PipelineJob
	if arg0 := args.Get(0); arg0 != nil {
		pipelineJob = arg0.(*aiplatformpb.PipelineJob)
	}

	return pipelineJob, args.Error(1)
}

func (m *MockPipelineJobClient) CreatePipelineJob(
	_ context.Context,
	req *aiplatformpb.CreatePipelineJobRequest,
	_ ...gax.CallOption,
) (*aiplatformpb.PipelineJob, error) {
	args := m.Called(req)
	var pipelineJob *aiplatformpb.PipelineJob
	if arg1 := args.Get(0); arg1 != nil {
		pipelineJob = arg1.(*aiplatformpb.PipelineJob)
	}
	return pipelineJob, args.Error(1)
}
