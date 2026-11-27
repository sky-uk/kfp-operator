//go:build unit

package mocks

import (
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/stretchr/testify/mock"
)

type MockPlatformService struct {
	mock.Mock
}

func (m *MockPlatformService) Apply(
	compiledPipeline resource.CompiledPipeline,
) (resource.CompiledPipeline, error) {
	args := m.Called(compiledPipeline)

	var result resource.CompiledPipeline
	if args.Get(0) != nil {
		result = args.Get(0).(resource.CompiledPipeline)
	}
	return result, args.Error(1)
}
