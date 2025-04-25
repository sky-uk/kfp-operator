package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockPipelineUploadService struct {
	mock.Mock
}

func (m *MockPipelineUploadService) UploadPipeline(
	ctx context.Context,
	content []byte,
	pipelineName string,
) (string, error) {
	args := m.Called(ctx, content, pipelineName)
	return args.String(0), args.Error(1)
}

func (m *MockPipelineUploadService) UploadPipelineVersion(
	ctx context.Context,
	id string,
	content []byte,
	version string,
) error {
	args := m.Called(ctx, id, content, version)
	return args.Error(0)
}
