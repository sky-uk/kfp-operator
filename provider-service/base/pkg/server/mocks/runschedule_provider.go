package mocks

import (
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/stretchr/testify/mock"
)

type MockRunScheduleProvider struct {
	mock.Mock
}

func (m *MockRunScheduleProvider) CreateRunSchedule(ed resource.RunScheduleDefinition) (string, error) {
	args := m.Called(ed)
	return args.String(0), args.Error(1)
}

func (m *MockRunScheduleProvider) UpdateRunSchedule(ed resource.RunScheduleDefinition, id string) (string, error) {
	args := m.Called(ed, id)
	return args.String(0), args.Error(1)
}

func (m *MockRunScheduleProvider) DeleteRunSchedule(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
