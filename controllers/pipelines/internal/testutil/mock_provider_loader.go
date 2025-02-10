//go:build unit

package testutil

import (
	"context"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/stretchr/testify/mock"
)

type MockProviderLoader struct {
	mock.Mock
}

func (m *MockProviderLoader) LoadProvider(_ context.Context, namespace string, name string) (pipelinesv1.Provider, error) {
	args := m.Called(namespace, name)
	var provider pipelinesv1.Provider
	if args.Get(0) != nil {
		provider = args.Get(0).(pipelinesv1.Provider)
	}
	return provider, args.Error(1)
}
