//go:build unit

package mocks

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/stretchr/testify/mock"
)

type MockLabelService struct {
	mock.Mock
}

func (m *MockLabelService) InsertLabelsIntoParameters(
	pipeline resource.CompiledPipeline,
	labels []string,
) (resource.CompiledPipeline, error) {
	args := m.Called(pipeline, labels)
	if args.Get(0) != nil {
		return args.Get(0).(resource.CompiledPipeline), args.Error(1)
	}
	return nil, fmt.Errorf("no return value set")
}
