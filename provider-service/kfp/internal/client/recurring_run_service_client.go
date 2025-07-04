package client

import (
	"context"

	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RecurringRunServiceClient interface {
	CreateRecurringRun(
		ctx context.Context,
		in *go_client.CreateRecurringRunRequest,
		opts ...grpc.CallOption,
	) (*go_client.RecurringRun, error)

	GetRecurringRun(
		ctx context.Context,
		in *go_client.GetRecurringRunRequest,
		opts ...grpc.CallOption,
	) (*go_client.RecurringRun, error)

	DeleteRecurringRun(
		ctx context.Context,
		in *go_client.DeleteRecurringRunRequest,
		opts ...grpc.CallOption,
	) (*emptypb.Empty, error)
}
