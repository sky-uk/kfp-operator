//go:build unit

package testutil

import (
	"context"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
)

type MockStatusManager struct {
	UpdateStatusFunc func() error
}

func (sm *MockStatusManager) UpdateProviderStatus(_ context.Context, _ *pipelinesv1.Provider, _ apis.SynchronizationState, _ string) error {
	if sm.UpdateStatusFunc != nil {
		return sm.UpdateStatusFunc()
	}
	return nil
}
