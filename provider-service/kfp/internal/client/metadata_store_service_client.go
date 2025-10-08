package client

import (
	"context"

	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/ml_metadata"
	"google.golang.org/grpc"
)

type MetadataStoreServiceClient interface {
	GetArtifactType(
		ctx context.Context,
		in *ml_metadata.GetArtifactTypeRequest,
		opts ...grpc.CallOption,
	) (*ml_metadata.GetArtifactTypeResponse, error)

	GetArtifactsByContext(
		ctx context.Context,
		in *ml_metadata.GetArtifactsByContextRequest,
		opts ...grpc.CallOption,
	) (*ml_metadata.GetArtifactsByContextResponse, error)

	GetContextByTypeAndName(
		ctx context.Context,
		in *ml_metadata.GetContextByTypeAndNameRequest,
		opts ...grpc.CallOption,
	) (*ml_metadata.GetContextByTypeAndNameResponse, error)

	GetExecutionsByContext(
		ctx context.Context,
		in *ml_metadata.GetExecutionsByContextRequest,
		opts ...grpc.CallOption,
	) (*ml_metadata.GetExecutionsByContextResponse, error)

	GetArtifactsByID(
		ctx context.Context,
		in *ml_metadata.GetArtifactsByIDRequest,
		opts ...grpc.CallOption,
	) (*ml_metadata.GetArtifactsByIDResponse, error)

	GetEventsByExecutionIDs(
		ctx context.Context,
		in *ml_metadata.GetEventsByExecutionIDsRequest,
		opts ...grpc.CallOption,
	) (*ml_metadata.GetEventsByExecutionIDsResponse, error)
}
