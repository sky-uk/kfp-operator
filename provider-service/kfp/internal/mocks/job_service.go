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
	experimentVersion string,
) (string, error) {
	args := m.Called(rsd, pipelineId, pipelineVersionId, experimentVersion)
	var res string
	if args.Get(0) != nil {
		res = args.String(0)
	}
	return res, args.Error(1)
}

func (m *MockJobService) GetJob(id string) (string, error) {
	args := m.Called(id)
	var res string
	if args.Get(0) != nil {
		res = args.String(0)
	}
	return res, args.Error(1)
}

func (m *MockJobService) DeleteJob(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
