//go:build unit

package mocks

import (
	"context"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/stretchr/testify/mock"
)

type MockStatusManager struct {
	mock.Mock
}

func (m *MockStatusManager) UpdateProviderStatus(_ context.Context, provider *pipelinesv1.Provider, state apis.SynchronizationState, message string) error {
	args := m.Called(provider, state, message)
	return args.Error(0)
}
