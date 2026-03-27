//go:build decoupled || unit

package mocks

import (
	"context"

	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/stretchr/testify/mock"
)

type MockMetadataStore struct {
	mock.Mock
}

func (m *MockMetadataStore) GetArtifactsForRun(
	_ context.Context,
	runId string,
) ([]common.PipelineComponent, error) {
	args := m.Called(runId)
	var pipelineComponents []common.PipelineComponent
	if arg0 := args.Get(0); arg0 != nil {
		pipelineComponents = arg0.([]common.PipelineComponent)
	}
	return pipelineComponents, args.Error(1)
}
