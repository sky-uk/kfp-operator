//go:build unit

package mocks

import (
	"fmt"
	"github.com/stretchr/testify/mock"
)

type MockLabelGen struct {
	mock.Mock
}

func (m *MockLabelGen) GenerateLabels(value any) (map[string]string, error) {
	args := m.Called(value)
	if args.Get(0) != nil {
		return args.Get(0).(map[string]string), args.Error(1)
	}
	return nil, fmt.Errorf("no return value set")
}
