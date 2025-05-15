package mocks

import (
	"context"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/mock"
)

type MockScheduleClient struct{ mock.Mock }

func (m *MockScheduleClient) CreateSchedule(
	_ context.Context,
	req *aiplatformpb.CreateScheduleRequest,
	_ ...gax.CallOption,
) (*aiplatformpb.Schedule, error) {
	args := m.Called(req)
	var schedule *aiplatformpb.Schedule
	if arg1 := args.Get(0); arg1 != nil {
		schedule = arg1.(*aiplatformpb.Schedule)
	}
	return schedule, args.Error(1)
}

func (m *MockScheduleClient) DeleteSchedule(
	_ context.Context,
	req *aiplatformpb.DeleteScheduleRequest,
	_ ...gax.CallOption,
) (*aiplatform.DeleteScheduleOperation, error) {
	args := m.Called(req)
	var operation *aiplatform.DeleteScheduleOperation
	if arg1 := args.Get(0); arg1 != nil {
		operation = arg1.(*aiplatform.DeleteScheduleOperation)
	}
	return operation, args.Error(1)
}
func (m *MockScheduleClient) UpdateSchedule(
	_ context.Context,
	req *aiplatformpb.UpdateScheduleRequest,
	_ ...gax.CallOption,
) (*aiplatformpb.Schedule, error) {
	args := m.Called(req)
	var schedule *aiplatformpb.Schedule
	if arg1 := args.Get(0); arg1 != nil {
		schedule = arg1.(*aiplatformpb.Schedule)
	}
	return schedule, args.Error(1)
}

func (m *MockScheduleClient) ListSchedules(
	_ context.Context,
	req *aiplatformpb.ListSchedulesRequest,
	_ ...gax.CallOption,
) *aiplatform.ScheduleIterator {
	args := m.Called(req)
	var scheduleIterator *aiplatform.ScheduleIterator
	if arg1 := args.Get(0); arg1 != nil {
		scheduleIterator = arg1.(*aiplatform.ScheduleIterator)
	}
	return scheduleIterator
}
