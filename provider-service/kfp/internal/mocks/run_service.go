package mocks

import (
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/stretchr/testify/mock"
)

type MockRunService struct {
	mock.Mock
}

func (m *MockRunService) CreateRun(
	rd resource.RunDefinition,
	pipelineId string,
	pipelineVersionId string,
	experimentVersion string,
) (string, error) {
	args := m.Called(rd, pipelineId, pipelineVersionId, experimentVersion)
	return args.String(0), args.Error(1)
}
