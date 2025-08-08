//go:build unit

package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockFileHandler struct{ mock.Mock }

func (m *MockFileHandler) Write(_ context.Context, p []byte, bucket string, filePath string) error {
	args := m.Called(p, bucket, filePath)
	return args.Error(0)
}

func (m *MockFileHandler) Delete(_ context.Context, id string, bucket string) error {
	args := m.Called(id, bucket)
	return args.Error(0)
}

func (m *MockFileHandler) Read(_ context.Context, bucket string, filePath string) (map[string]any, error) {
	args := m.Called(bucket, filePath)
	var data map[string]any
	if arg0 := args.Get(0); arg0 != nil {
		data = arg0.(map[string]any)
	}
	return data, args.Error(1)
}
