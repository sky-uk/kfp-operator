//go:build unit

package mocks

import (
	"context"

	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockExperimentServiceClient struct {
	mock.Mock
}

func (m *MockExperimentServiceClient) CreateExperiment(
	_ context.Context,
	in *go_client.CreateExperimentRequest,
	_ ...grpc.CallOption,
) (*go_client.Experiment, error) {
	args := m.Called(in)
	var exp *go_client.Experiment
	if arg0 := args.Get(0); arg0 != nil {
		exp = arg0.(*go_client.Experiment)
	}
	return exp, args.Error(1)
}

func (m *MockExperimentServiceClient) ListExperiments(
	_ context.Context,
	in *go_client.ListExperimentsRequest,
	_ ...grpc.CallOption,
) (*go_client.ListExperimentsResponse, error) {
	args := m.Called(in)
	var res *go_client.ListExperimentsResponse
	if arg0 := args.Get(0); arg0 != nil {
		res = arg0.(*go_client.ListExperimentsResponse)
	}
	return res, args.Error(1)
}

func (m *MockExperimentServiceClient) DeleteExperiment(
	_ context.Context,
	in *go_client.DeleteExperimentRequest,
	_ ...grpc.CallOption,
) (*emptypb.Empty, error) {
	args := m.Called(in)
	return &emptypb.Empty{}, args.Error(0)
}
