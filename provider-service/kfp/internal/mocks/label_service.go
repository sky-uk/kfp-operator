//go:build unit

package mocks

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/mock"
)

type MockLabelService struct {
	mock.Mock
}

func (m *MockLabelService) InsertLabelsIntoParameters(jsonBytes []byte, labels []string) ([]byte, error) {
	args := m.Called(jsonBytes, labels)
	if args.Get(0) != nil {
		return args.Get(0).(json.RawMessage), args.Error(1)
	}
	return nil, fmt.Errorf("no return value set")
}
