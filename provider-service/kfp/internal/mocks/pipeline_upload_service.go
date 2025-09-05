//go:build unit

package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockPipelineUploadService struct {
	mock.Mock
}

func (m *MockPipelineUploadService) UploadPipeline(
	_ context.Context,
	content []byte,
	pipelineName string,
) (string, error) {
	args := m.Called(content, pipelineName)
	return args.String(0), args.Error(1)
}

func (m *MockPipelineUploadService) UploadPipelineVersion(
	_ context.Context,
	id string,
	content []byte,
	version string,
) error {
	args := m.Called(id, content, version)
	return args.Error(0)
}
