//go:build unit

package mocks

import (
	"context"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/stretchr/testify/mock"
)

type MockProviderStatusUpdater struct {
	mock.Mock
}

func (m *MockProviderStatusUpdater) UpdateStatus(
	_ context.Context,
	provider *pipelineshub.Provider,
	state apis.SynchronizationState,
	message string,
) error {
	args := m.Called(provider, state, message)
	return args.Error(0)
}
