package mocks

import (
	"context"
	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/stretchr/testify/mock"
)

type MockExperimentService struct {
	mock.Mock
}

func (m *MockExperimentService) CreateExperiment(
	_ context.Context,
	experiment common.NamespacedName,
	description string,
) (string, error) {
	args := m.Called(experiment, description)
	return args.String(0), args.Error(1)
}

func (m *MockExperimentService) DeleteExperiment(
	_ context.Context,
	id string,
) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockExperimentService) ExperimentIdByDisplayName(
	_ context.Context,
	experiment common.NamespacedName,
) (string, error) {
	args := m.Called(experiment)
	return args.String(0), args.Error(1)
}
