package mocks

import (
	"context"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/stretchr/testify/mock"
)

type MockRunService struct {
	mock.Mock
}

func (m *MockRunService) CreateRun(
	ctx context.Context,
	rd resource.RunDefinition,
	pipelineId string,
	pipelineVersionId string,
	experimentId string,
) (string, error) {
	args := m.Called(ctx, rd, pipelineId, pipelineVersionId, experimentId)
	return args.String(0), args.Error(1)
}
