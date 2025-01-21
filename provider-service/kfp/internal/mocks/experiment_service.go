package mocks

import (
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/stretchr/testify/mock"
)

type MockExperimentService struct {
	mock.Mock
}

func (m *MockExperimentService) CreateExperiment(
	experiment common.NamespacedName,
	description string,
) (string, error) {
	args := m.Called(experiment, description)
	var res string
	if args.Get(0) != nil {
		res = args.String(0)
	}
	return res, args.Error(1)
}

func (m *MockExperimentService) DeleteExperiment(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockExperimentService) ExperimentIdByName(
	experiment common.NamespacedName,
) (string, error) {
	args := m.Called(experiment)
	var res string
	if args.Get(0) != nil {
		res = args.String(0)
	}
	return res, args.Error(1)
}
