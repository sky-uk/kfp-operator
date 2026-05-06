//go:build unit

package mocks

import (
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/sky-uk/kfp-operator/pkg/providers/base"
	"github.com/stretchr/testify/mock"
)

type MockJobBuilder struct{ mock.Mock }

func (m *MockJobBuilder) MkRunPipelineJob(
	rd base.RunDefinition,
	networkAttachment string,
) (*aiplatformpb.PipelineJob, error) {
	args := m.Called(rd, networkAttachment)
	var pipelineJob *aiplatformpb.PipelineJob
	if arg0 := args.Get(0); arg0 != nil {
		pipelineJob = arg0.(*aiplatformpb.PipelineJob)
	}
	return pipelineJob, args.Error(1)
}

func (m *MockJobBuilder) MkRunSchedulePipelineJob(
	rsd base.RunScheduleDefinition,
	networkAttachment string,
) (*aiplatformpb.PipelineJob, error) {
	args := m.Called(rsd, networkAttachment)
	var pipelineJob *aiplatformpb.PipelineJob
	if arg0 := args.Get(0); arg0 != nil {
		pipelineJob = arg0.(*aiplatformpb.PipelineJob)
	}
	return pipelineJob, args.Error(1)
}

func (m *MockJobBuilder) MkSchedule(
	rsd base.RunScheduleDefinition,
	pipelineJob *aiplatformpb.PipelineJob,
	parent string,
	maxConcurrentRunCount int64,
) (*aiplatformpb.Schedule, error) {
	args := m.Called(rsd, pipelineJob, parent, maxConcurrentRunCount)
	var schedule *aiplatformpb.Schedule
	if arg0 := args.Get(0); arg0 != nil {
		schedule = arg0.(*aiplatformpb.Schedule)
	}
	return schedule, args.Error(1)
}
