package mocks

import (
	context "context"

	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	"github.com/stretchr/testify/mock"
	grpc "google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockRecurringRunServiceClient struct {
	mock.Mock
}

func (m *MockRecurringRunServiceClient) CreateRecurringRun(
	ctx context.Context,
	in *go_client.CreateRecurringRunRequest,
	opts ...grpc.CallOption,
) (*go_client.RecurringRun, error) {
	args := m.Called(in)
	var recurringRun *go_client.RecurringRun
	if arg0 := args.Get(0); arg0 != nil {
		recurringRun = arg0.(*go_client.RecurringRun)
	}
	return recurringRun, args.Error(1)
}

func (m *MockRecurringRunServiceClient) GetRecurringRun(
	ctx context.Context,
	in *go_client.GetRecurringRunRequest,
	opts ...grpc.CallOption,
) (*go_client.RecurringRun, error) {
	args := m.Called(in)
	var recurringRun *go_client.RecurringRun
	if arg0 := args.Get(0); arg0 != nil {
		recurringRun = arg0.(*go_client.RecurringRun)
	}
	return recurringRun, args.Error(1)
}

func (m *MockRecurringRunServiceClient) DeleteRecurringRun(
	ctx context.Context,
	in *go_client.DeleteRecurringRunRequest,
	opts ...grpc.CallOption,
) (*emptypb.Empty, error) {
	args := m.Called(in)
	return &emptypb.Empty{}, args.Error(0)
}
