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
	_ []byte,
	pipelineName string,
) (string, error) {
	args := m.Called(pipelineName)
	return args.String(0), args.Error(1)
}

func (m *MockPipelineUploadService) UploadPipelineVersion(
	_ context.Context,
	id string,
	_ []byte,
	version string,
) error {
	args := m.Called(id, version)
	return args.Error(0)
}
