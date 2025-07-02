package mocks

import (
	"context"
	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockPipelineServiceClient struct {
	mock.Mock
}

func (m *MockPipelineServiceClient) DeletePipeline(
	_ context.Context,
	in *go_client.DeletePipelineRequest,
	_ ...grpc.CallOption,
) (*emptypb.Empty, error) {
	args := m.Called(in)
	return &emptypb.Empty{}, args.Error(0)
}

func (m *MockPipelineServiceClient) ListPipelines(
	_ context.Context,
	in *go_client.ListPipelinesRequest,
	_ ...grpc.CallOption,
) (*go_client.ListPipelinesResponse, error) {
	args := m.Called(in)
	var res *go_client.ListPipelinesResponse
	if arg0 := args.Get(0); arg0 != nil {
		res = arg0.(*go_client.ListPipelinesResponse)
	}
	return res, args.Error(1)
}

func (m *MockPipelineServiceClient) ListPipelineVersions(
	_ context.Context,
	in *go_client.ListPipelineVersionsRequest,
	_ ...grpc.CallOption,
) (*go_client.ListPipelineVersionsResponse, error) {
	args := m.Called(in)
	var res *go_client.ListPipelineVersionsResponse
	if arg0 := args.Get(0); arg0 != nil {
		res = arg0.(*go_client.ListPipelineVersionsResponse)
	}
	return res, args.Error(1)
}
