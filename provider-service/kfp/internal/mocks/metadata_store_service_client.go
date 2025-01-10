package mocks

import (
	context "context"

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
