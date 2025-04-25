package mocks

import (
	"context"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/stretchr/testify/mock"
)

type MockExperimentService struct {
	mock.Mock
}

func (m *MockExperimentService) CreateExperiment(
	ctx context.Context,
	experiment common.NamespacedName,
	description string,
) (string, error) {
	args := m.Called(ctx, experiment, description)
	return args.String(0), args.Error(1)
}

func (m *MockExperimentService) DeleteExperiment(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockExperimentService) ExperimentIdByName(
	ctx context.Context,
	experiment common.NamespacedName,
) (string, error) {
	args := m.Called(ctx, experiment)
	return args.String(0), args.Error(1)
}
