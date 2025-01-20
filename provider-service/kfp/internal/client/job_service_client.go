package client

import (
	"context"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type JobServiceClient interface {
	CreateJob(
		ctx context.Context,
		in *go_client.CreateJobRequest,
		opts ...grpc.CallOption,
	) (*go_client.Job, error)

	GetJob(
		ctx context.Context,
		in *go_client.GetJobRequest,
		opts ...grpc.CallOption,
	) (*go_client.Job, error)

	DeleteJob(
		ctx context.Context,
		in *go_client.DeleteJobRequest,
		opts ...grpc.CallOption,
	) (*emptypb.Empty, error)
}
