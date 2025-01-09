package mocks

import (
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/stretchr/testify/mock"
)

type MockRunScheduleProvider struct {
	mock.Mock
}

func (m *MockRunScheduleProvider) CreateRunSchedule(
	rsd resource.RunScheduleDefinition,
) (string, error) {
	args := m.Called(rsd)
	return args.String(0), args.Error(1)
}

func (m *MockRunScheduleProvider) UpdateRunSchedule(
	rsd resource.RunScheduleDefinition,
	id string,
) (string, error) {
	args := m.Called(rsd, id)
	return args.String(0), args.Error(1)
}

func (m *MockRunScheduleProvider) DeleteRunSchedule(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
