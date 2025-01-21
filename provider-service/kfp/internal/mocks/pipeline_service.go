package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockPipelineService struct {
	mock.Mock
}

func (m *MockPipelineService) DeletePipeline(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPipelineService) PipelineIdForName(pipelineName string) (string, error) {
	args := m.Called(pipelineName)
	var res string
	if args.Get(0) != nil {
		res = args.Get(0).(string)
	}
	return res, args.Error(1)
}

func (m *MockPipelineService) PipelineVersionIdForName(versionName string, pipelineId string) (string, error) {
	args := m.Called(versionName, pipelineId)
	var res string
	if args.Get(0) != nil {
		res = args.Get(0).(string)
	}
	return res, args.Error(1)
}
