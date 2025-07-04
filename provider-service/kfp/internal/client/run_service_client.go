package client

import (
	"context"

	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	"google.golang.org/grpc"
)

type RunServiceClient interface {
	GetRun(
		ctx context.Context,
		in *go_client.GetRunRequest,
		opts ...grpc.CallOption,
	) (*go_client.Run, error)

	CreateRun(
		ctx context.Context,
		in *go_client.CreateRunRequest,
		opts ...grpc.CallOption,
	) (*go_client.Run, error)
}
