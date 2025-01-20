package client // Creates a new experiment.

import (
	"context"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ExperimentServiceClient interface {
	CreateExperiment(
		ctx context.Context,
		in *go_client.CreateExperimentRequest,
		opts ...grpc.CallOption,
	) (*go_client.Experiment, error)

	DeleteExperiment(
		ctx context.Context,
		in *go_client.DeleteExperimentRequest,
		opts ...grpc.CallOption,
	) (*emptypb.Empty, error)

	ListExperiment(
		ctx context.Context,
		in *go_client.ListExperimentsRequest,
		opts ...grpc.CallOption,
	) (*go_client.ListExperimentsResponse, error)
}
