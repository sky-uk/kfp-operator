package mocks

import (
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/stretchr/testify/mock"
)

type MockJobService struct {
	mock.Mock
}

func (m *MockJobService) CreateJob(rsd resource.RunScheduleDefinition,
	pipelineId string,
	pipelineVersionId string,
	experimentId string,
) (string, error) {
	args := m.Called(rsd, pipelineId, pipelineVersionId, experimentId)
	return args.String(0), args.Error(1)
}

func (m *MockJobService) GetJob(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *MockJobService) DeleteJob(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
