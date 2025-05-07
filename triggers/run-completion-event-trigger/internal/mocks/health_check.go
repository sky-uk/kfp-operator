//go:build unit

package mocks

import "github.com/stretchr/testify/mock"

type MockHealthCheck struct {
	mock.Mock
}

func (m *MockHealthCheck) IsHealthy() bool {
	args := m.Called()
	return args.Bool(0)
}
