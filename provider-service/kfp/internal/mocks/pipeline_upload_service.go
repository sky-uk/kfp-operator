package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockPipelineUploadService struct {
	mock.Mock
}

func (m *MockPipelineUploadService) UploadPipeline(
	content []byte,
	pipelineName string,
) (string, error) {
	args := m.Called(content, pipelineName)
	return args.String(0), args.Error(1)
}

func (m *MockPipelineUploadService) UploadPipelineVersion(
	id string,
	content []byte,
	version string,
) error {
	args := m.Called(id, content, version)
	return args.Error(0)
}
