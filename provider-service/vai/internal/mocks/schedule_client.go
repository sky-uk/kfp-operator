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
	ctx context.Context,
	req *aiplatformpb.CreateScheduleRequest,
	opts ...gax.CallOption,
) (*aiplatformpb.Schedule, error) {
	args := m.Called(ctx, req, opts)
	var schedule *aiplatformpb.Schedule
	if arg1 := args.Get(0); arg1 != nil {
		schedule = arg1.(*aiplatformpb.Schedule)
	}
	return schedule, args.Error(1)
}

func (m *MockScheduleClient) DeleteSchedule(
	ctx context.Context,
	req *aiplatformpb.DeleteScheduleRequest,
	opts ...gax.CallOption,
) (*aiplatform.DeleteScheduleOperation, error) {
	args := m.Called(ctx, req, opts)
	var operation *aiplatform.DeleteScheduleOperation
	if arg1 := args.Get(0); arg1 != nil {
		operation = arg1.(*aiplatform.DeleteScheduleOperation)
	}
	return operation, args.Error(1)
}
func (m *MockScheduleClient) UpdateSchedule(
	ctx context.Context,
	req *aiplatformpb.UpdateScheduleRequest,
	opts ...gax.CallOption,
) (*aiplatformpb.Schedule, error) {
	args := m.Called(ctx, req, opts)
	var schedule *aiplatformpb.Schedule
	if arg1 := args.Get(0); arg1 != nil {
		schedule = arg1.(*aiplatformpb.Schedule)
	}
	return schedule, args.Error(1)
}
