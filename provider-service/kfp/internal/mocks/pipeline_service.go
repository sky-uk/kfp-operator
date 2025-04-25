package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockPipelineService struct {
	mock.Mock
}

func (m *MockPipelineService) DeletePipeline(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPipelineService) PipelineIdForName(
	ctx context.Context,
	pipelineName string,
) (string, error) {
	args := m.Called(ctx, pipelineName)
	return args.String(0), args.Error(1)
}

func (m *MockPipelineService) PipelineVersionIdForName(
	ctx context.Context,
	versionName string,
	pipelineId string,
) (string, error) {
	args := m.Called(ctx, versionName, pipelineId)
	return args.String(0), args.Error(1)
}
