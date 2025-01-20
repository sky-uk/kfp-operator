package client

import (
	"context"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/googleapis/gax-go/v2"
)

type ScheduleClient interface {
	CreateSchedule(
		ctx context.Context,
		req *aiplatformpb.CreateScheduleRequest,
		opts ...gax.CallOption,
	) (*aiplatformpb.Schedule, error)

	DeleteSchedule(
		ctx context.Context,
		req *aiplatformpb.DeleteScheduleRequest,
		opts ...gax.CallOption,
	) (*aiplatform.DeleteScheduleOperation, error)

	UpdateSchedule(
		ctx context.Context,
		req *aiplatformpb.UpdateScheduleRequest,
		opts ...gax.CallOption,
	) (*aiplatformpb.Schedule, error)
}
