package provider

import (
	"context"
	"fmt"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/provider/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PipelineService interface {
	DeletePipeline(ctx context.Context, id string) error
	PipelineIdForName(ctx context.Context, pipelineName string) (string, error)
	PipelineVersionIdForName(ctx context.Context, versionName string, pipelineId string) (string, error)
}

type DefaultPipelineService struct {
	client client.PipelineServiceClient
}

func NewPipelineService(
	conn *grpc.ClientConn,
) (PipelineService, error) {
	if conn == nil {
		return nil, fmt.Errorf(
			"no gRPC connection was provided to start pipeline service",
		)
	}

	return &DefaultPipelineService{
		client: go_client.NewPipelineServiceClient(conn),
	}, nil
}

// DeletePipeline delete a pipline by pipeline id. Does no error if there is no
// such pipeline id.
func (ps *DefaultPipelineService) DeletePipeline(ctx context.Context, id string) error {
	_, err := ps.client.DeletePipeline(
		ctx,
		&go_client.DeletePipelineRequest{
			Id: id,
		},
	)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			// is a gRPC error
			switch st.Code() {
			case codes.NotFound:
				return nil
			default:
				return err
			}
		}
		return err
	}

	return nil
}

// PipelineIdForName gets the pipeline id corresponding to the pipeline name.
// Expects to find exactly one such pipeline.
func (ps *DefaultPipelineService) PipelineIdForName(
	ctx context.Context,
	pipelineName string,
) (string, error) {
	res, err := ps.client.ListPipelines(
		ctx,
		&go_client.ListPipelinesRequest{
			Filter: *util.ByNameFilter(pipelineName),
		},
	)
	if err != nil {
		return "", err
	}
	pipelineCount := len(res.Pipelines)
	if pipelineCount != 1 {
		return "", fmt.Errorf("found %d pipelines, expected exactly one", pipelineCount)
	}

	return res.Pipelines[0].Id, nil
}

// PipelineVersionIdForName gets the pipeline version corresponding to the
// pipeline id. Expects to find exactly one such pipeline.
func (ps *DefaultPipelineService) PipelineVersionIdForName(
	ctx context.Context,
	versionName string,
	pipelineId string,
) (string, error) {
	res, err := ps.client.ListPipelineVersions(
		ctx,
		&go_client.ListPipelineVersionsRequest{
			ResourceKey: &go_client.ResourceKey{
				Id: pipelineId,
			},
			Filter: *util.ByNameFilter(versionName),
		},
	)
	if err != nil {
		return "", err
	}

	pipelineVersionCount := len(res.Versions)
	if pipelineVersionCount != 1 {
		return "", fmt.Errorf(
			"found %d pipeline versions, expected exactly one",
			pipelineVersionCount,
		)
	}
	return res.Versions[0].Id, nil
}
