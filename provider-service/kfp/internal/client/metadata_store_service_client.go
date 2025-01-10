package client

import (
	"context"

	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/ml_metadata"
	"google.golang.org/grpc"
)

type MetadataStoreServiceClient interface {
	GetArtifactType(
		_ context.Context,
		in *ml_metadata.GetArtifactTypeRequest,
		_ ...grpc.CallOption,
	) (*ml_metadata.GetArtifactTypeResponse, error)

	GetArtifactsByContext(
		_ context.Context,
		in *ml_metadata.GetArtifactsByContextRequest,
		_ ...grpc.CallOption,
	) (*ml_metadata.GetArtifactsByContextResponse, error)

	GetContextByTypeAndName(
		_ context.Context,
		in *ml_metadata.GetContextByTypeAndNameRequest,
		_ ...grpc.CallOption,
	) (*ml_metadata.GetContextByTypeAndNameResponse, error)
}
