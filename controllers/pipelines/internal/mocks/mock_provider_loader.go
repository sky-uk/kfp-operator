//go:build unit

package mocks

import (
	"context"

	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/stretchr/testify/mock"
)

type MockProviderLoader struct {
	mock.Mock
}

func (m *MockProviderLoader) LoadProvider(
	_ context.Context,
	desiredProvider common.NamespacedName,
) (pipelineshub.Provider, error) {
	args := m.Called(desiredProvider)
	var provider pipelineshub.Provider
	if args.Get(0) != nil {
		provider = args.Get(0).(pipelineshub.Provider)
	}
	return provider, args.Error(1)
}
