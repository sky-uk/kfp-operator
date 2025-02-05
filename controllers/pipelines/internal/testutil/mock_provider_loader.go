//go:build unit

package testutil

import (
	"context"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
)

type MockProviderLoader struct {
	LoadProviderFunc func() (pipelinesv1.Provider, error)
}

func (m MockProviderLoader) LoadProvider(_ context.Context, _ string, _ string) (pipelinesv1.Provider, error) {
	if m.LoadProviderFunc != nil {
		return m.LoadProviderFunc()
	}
	return pipelinesv1.Provider{}, nil
}
