package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockPipelineService struct {
	mock.Mock
}

func (m *MockPipelineService) DeletePipeline(
	_ context.Context,
	id string,
) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPipelineService) PipelineIdForDisplayName(
	_ context.Context,
	pipelineName string,
) (string, error) {
	args := m.Called(pipelineName)
	return args.String(0), args.Error(1)
}

func (m *MockPipelineService) PipelineVersionIdForDisplayName(
	_ context.Context,
	versionName string,
	pipelineId string,
) (string, error) {
	args := m.Called(versionName, pipelineId)
	return args.String(0), args.Error(1)
}
