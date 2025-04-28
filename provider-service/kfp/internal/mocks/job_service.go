package mocks

import (
	"context"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/stretchr/testify/mock"
)

type MockJobService struct {
	mock.Mock
}

func (m *MockJobService) CreateJob(
	_ context.Context,
	rsd resource.RunScheduleDefinition,
	pipelineId string,
	pipelineVersionId string,
	experimentId string,
) (string, error) {
	args := m.Called(rsd, pipelineId, pipelineVersionId, experimentId)
	return args.String(0), args.Error(1)
}

func (m *MockJobService) GetJob(_ context.Context, id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *MockJobService) DeleteJob(_ context.Context, id string) error {
	args := m.Called(id)
	return args.Error(0)
}
