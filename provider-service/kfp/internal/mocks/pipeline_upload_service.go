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
	filePath string,
) (string, error) {
	args := m.Called(content, pipelineName, filePath)
	var res string
	if arg0 := args.Get(0); arg0 != nil {
		res = arg0.(string)
	}
	return res, args.Error(1)
}

func (m *MockPipelineUploadService) UploadPipelineVersion(
	id string,
	content []byte,
	version string,
	filePath string,
) error {
	args := m.Called(id, content, version, filePath)
	return args.Error(0)
}
