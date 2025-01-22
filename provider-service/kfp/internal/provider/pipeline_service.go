package provider

import (
	"context"
	"fmt"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PipelineService interface {
	DeletePipeline(id string) error
	PipelineIdForName(pipelineName string) (string, error)
	PipelineVersionIdForName(versionName string, pipelineId string) (string, error)
}

type DefaultPipelineService struct {
	ctx    context.Context
	client client.PipelineServiceClient
}

func NewPipelineService(
	ctx context.Context,
	conn *grpc.ClientConn,
) (PipelineService, error) {
	if conn == nil {
		return nil, fmt.Errorf(
			"no gRPC connection was provided to start pipeline service",
		)
	}

	return &DefaultPipelineService{
		ctx:    ctx,
		client: go_client.NewPipelineServiceClient(conn),
	}, nil
}

func (ps *DefaultPipelineService) DeletePipeline(id string) error {
	_, err := ps.client.DeletePipeline(
		ps.ctx,
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

func (ps *DefaultPipelineService) PipelineIdForName(
	pipelineName string,
) (string, error) {
	res, err := ps.client.ListPipelines(
		ps.ctx,
		&go_client.ListPipelinesRequest{
			Filter: *byNameFilter(pipelineName),
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

func (ps *DefaultPipelineService) PipelineVersionIdForName(
	versionName string,
	pipelineId string,
) (string, error) {
	res, err := ps.client.ListPipelineVersions(
		ps.ctx,
		&go_client.ListPipelineVersionsRequest{
			ResourceKey: &go_client.ResourceKey{
				Id: pipelineId,
			},
			Filter: *pipelineVersionByNameFilter(versionName),
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

// TODO: check if string_value works instead of stringValue because it is
// different from original implementation
// https://github.com/kubeflow/pipelines/blob/master/backend/api/v1beta1/filter.proto
func byNameFilter(name string) *string {
	filter := fmt.Sprintf(
		`{"predicates": [{"op": EQUALS, "key": "name", "string_value": "%s"}]}`,
		name,
	)
	return &filter
}

func pipelineVersionByNameFilter(name string) *string {
	filter := fmt.Sprintf(
		`{"predicates": [{"op": "EQUALS", "key": "name", "string_value": "%s"}]}`,
		name,
	)
	return &filter
}
