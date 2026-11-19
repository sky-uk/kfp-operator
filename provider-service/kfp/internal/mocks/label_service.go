//go:build unit

package mocks

import (
	"github.com/kubeflow/pipelines/api/v2alpha1/go/pipelinespec"
	"github.com/stretchr/testify/mock"
)

type MockLabelService struct {
	mock.Mock
}

func (m *MockLabelService) InsertLabelsIntoParameters(
	pipelineSpec *pipelinespec.PipelineSpec,
	labels []string,
) {
	m.Called(pipelineSpec, labels)
}
