//go:build unit

package mocks

import (
	"context"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/stretchr/testify/mock"
)

type MockStatusManager struct {
	mock.Mock
}

func (m *MockStatusManager) UpdateProviderStatus(_ context.Context, provider *pipelineshub.Provider, state apis.SynchronizationState, message string) error {
	args := m.Called(provider, state, message)
	return args.Error(0)
}
