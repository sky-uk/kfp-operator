package mocks

import (
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/stretchr/testify/mock"
)

type MockRunProvider struct {
	mock.Mock
}

func (m *MockRunProvider) CreateRun(ed resource.RunDefinition) (string, error) {
	args := m.Called(ed)
	return args.String(0), args.Error(1)
}

func (m *MockRunProvider) UpdateRun(ed resource.RunDefinition, id string) (string, error) {
	args := m.Called(ed, id)
	return args.String(0), args.Error(1)
}

func (m *MockRunProvider) DeleteRun(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
