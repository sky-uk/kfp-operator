package mocks

import (
	context "context"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/stretchr/testify/mock"
	grpc "google.golang.org/grpc"
)

type MockJobServiceClient struct {
	mock.Mock
}

func (m *MockJobServiceClient) GetJob(
	_ context.Context,
	in *go_client.GetJobRequest,
	_ ...grpc.CallOption,
) (*go_client.Job, error) {
	args := m.Called(in)
	var job *go_client.Job
	if arg0 := args.Get(0); arg0 != nil {
		job = arg0.(*go_client.Job)
	}
	return job, args.Error(1)
}

func (m *MockJobServiceClient) CreateJob(
	_ context.Context,
	in *go_client.CreateJobRequest,
	_ ...grpc.CallOption,
) (*go_client.Job, error) {
	args := m.Called(in)
	var job *go_client.Job
	if arg0 := args.Get(0); arg0 != nil {
		job = arg0.(*go_client.Job)
	}
	return job, args.Error(1)
}
