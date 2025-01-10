package mocks

import (
	context "context"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/stretchr/testify/mock"
	grpc "google.golang.org/grpc"
)

type MockRunServiceClient struct {
	mock.Mock
}

func (m *MockRunServiceClient) GetRun(
	_ context.Context,
	in *go_client.GetRunRequest,
	_ ...grpc.CallOption,
) (*go_client.RunDetail, error) {
	args := m.Called(in)
	var rd *go_client.RunDetail
	if arg0 := args.Get(0); arg0 != nil {
		rd = arg0.(*go_client.RunDetail)
	}
	return rd, args.Error(1)
}
