package mocks

import (
	"context"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/ml_metadata"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockMetadataStoreServiceClient struct {
	mock.Mock
}

func (m *MockMetadataStoreServiceClient) GetArtifactType(
	_ context.Context,
	in *ml_metadata.GetArtifactTypeRequest,
	_ ...grpc.CallOption,
) (*ml_metadata.GetArtifactTypeResponse, error) {
	args := m.Called(in)
	var response *ml_metadata.GetArtifactTypeResponse
	if arg0 := args.Get(0); arg0 != nil {
		response = arg0.(*ml_metadata.GetArtifactTypeResponse)
	}
	return response, args.Error(1)
}

func (m *MockMetadataStoreServiceClient) GetArtifactsByContext(
	_ context.Context,
	in *ml_metadata.GetArtifactsByContextRequest,
	_ ...grpc.CallOption,
) (*ml_metadata.GetArtifactsByContextResponse, error) {
	args := m.Called(in)
	var response *ml_metadata.GetArtifactsByContextResponse
	if arg0 := args.Get(0); arg0 != nil {
		response = arg0.(*ml_metadata.GetArtifactsByContextResponse)
	}
	return response, args.Error(1)
}

func (m *MockMetadataStoreServiceClient) GetContextByTypeAndName(
	_ context.Context,
	in *ml_metadata.GetContextByTypeAndNameRequest,
	_ ...grpc.CallOption,
) (*ml_metadata.GetContextByTypeAndNameResponse, error) {
	args := m.Called(in)
	var response *ml_metadata.GetContextByTypeAndNameResponse
	if arg0 := args.Get(0); arg0 != nil {
		response = arg0.(*ml_metadata.GetContextByTypeAndNameResponse)
	}
	return response, args.Error(1)
}

func (m *MockMetadataStoreServiceClient) GetArtifacts(
	_ context.Context, in *ml_metadata.GetArtifactsRequest, _ ...grpc.CallOption,
) (*ml_metadata.GetArtifactsResponse, error) {
	args := m.Called(in)
	var response *ml_metadata.GetArtifactsResponse
	if arg0 := args.Get(0); arg0 != nil {
		response = arg0.(*ml_metadata.GetArtifactsResponse)
	}
	return response, args.Error(1)
}

func (m *MockMetadataStoreServiceClient) GetEventsByExecutionIDs(
	_ context.Context, in *ml_metadata.GetEventsByExecutionIDsRequest, _ ...grpc.CallOption,
) (*ml_metadata.GetEventsByExecutionIDsResponse, error) {
	args := m.Called(in)
	var response *ml_metadata.GetEventsByExecutionIDsResponse
	if arg0 := args.Get(0); arg0 != nil {
		response = arg0.(*ml_metadata.GetEventsByExecutionIDsResponse)
	}
	return response, args.Error(1)
}

func (m *MockMetadataStoreServiceClient) GetArtifactsByID(
	_ context.Context,
	in *ml_metadata.GetArtifactsByIDRequest,
	_ ...grpc.CallOption,
) (*ml_metadata.GetArtifactsByIDResponse, error) {
	args := m.Called(in)
	var response *ml_metadata.GetArtifactsByIDResponse
	if arg0 := args.Get(0); arg0 != nil {
		response = arg0.(*ml_metadata.GetArtifactsByIDResponse)
	}
	return response, args.Error(1)
}

func (m *MockMetadataStoreServiceClient) GetExecutionsByContext(
	_ context.Context,
	in *ml_metadata.GetExecutionsByContextRequest,
	_ ...grpc.CallOption,
) (*ml_metadata.GetExecutionsByContextResponse, error) {
	args := m.Called(in)
	var response *ml_metadata.GetExecutionsByContextResponse
	if arg0 := args.Get(0); arg0 != nil {
		response = arg0.(*ml_metadata.GetExecutionsByContextResponse)
	}
	return response, args.Error(1)
}
