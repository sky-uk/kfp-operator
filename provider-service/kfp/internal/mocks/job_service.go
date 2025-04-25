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
	ctx context.Context,
	rsd resource.RunScheduleDefinition,
	pipelineId string,
	pipelineVersionId string,
	experimentId string,
) (string, error) {
	args := m.Called(ctx, rsd, pipelineId, pipelineVersionId, experimentId)
	return args.String(0), args.Error(1)
}

func (m *MockJobService) GetJob(ctx context.Context, id string) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

func (m *MockJobService) DeleteJob(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
