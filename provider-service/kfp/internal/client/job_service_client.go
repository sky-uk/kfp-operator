package client

import (
	"context"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	"google.golang.org/grpc"
)

type JobServiceClient interface {
	GetJob(
		ctx context.Context,
		in *go_client.GetJobRequest,
		opts ...grpc.CallOption,
	) (*go_client.Job, error)

	CreateJob(
		ctx context.Context,
		in *go_client.CreateJobRequest,
		opts ...grpc.CallOption,
	) (*go_client.Job, error)
}
