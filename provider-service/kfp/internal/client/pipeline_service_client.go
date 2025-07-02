package client

import (
	"context"
	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PipelineServiceClient interface {
	DeletePipeline(
		ctx context.Context,
		in *go_client.DeletePipelineRequest,
		opts ...grpc.CallOption,
	) (*emptypb.Empty, error)

	ListPipelines(
		ctx context.Context,
		in *go_client.ListPipelinesRequest,
		opts ...grpc.CallOption,
	) (*go_client.ListPipelinesResponse, error)

	ListPipelineVersions(
		ctx context.Context,
		in *go_client.ListPipelineVersionsRequest,
		opts ...grpc.CallOption,
	) (*go_client.ListPipelineVersionsResponse, error)
}
