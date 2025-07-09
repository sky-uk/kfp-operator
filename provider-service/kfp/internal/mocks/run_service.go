package mocks

import (
	"context"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/stretchr/testify/mock"
)

type MockRunService struct {
	mock.Mock
}

func (m *MockRunService) CreateRun(
	_ context.Context,
	rd base.RunDefinition,
	pipelineId string,
	pipelineVersionId string,
	experimentId string,
) (string, error) {
	args := m.Called(rd, pipelineId, pipelineVersionId, experimentId)
	return args.String(0), args.Error(1)
}
