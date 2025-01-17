package provider

import (
	"context"
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_client/pipeline_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_client/pipeline_upload_service"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"net/url"
)

func byNameFilter(name string) *string {
	filter := fmt.Sprintf(`{"predicates": [{"op": 1, "key": "name", "stringValue": "%s"}]}`, name)
	return &filter
}

// ListPipelineVersions uses a different filter syntax than other List* operations
func pipelineVersionByNameFilter(name string) *string {
	filter := fmt.Sprintf(`{"predicates": [{"op": "EQUALS", "key": "name", "string_value": "%s"}]}`, name)
	return &filter
}

type PipelineService struct {
	*pipeline_service.Client
	context context.Context
}

func NewPipelineService(ctx context.Context, providerConfig config.KfpProviderConfig) (*PipelineService, error) {
	apiUrl, err := url.Parse(providerConfig.Parameters.RestKfpApiUrl)
	if err != nil {
		return nil, err
	}

	return &PipelineService{
		Client: pipeline_client.NewHTTPClientWithConfig(strfmt.NewFormats(), &pipeline_client.TransportConfig{
			Host:     apiUrl.Host,
			Schemes:  []string{apiUrl.Scheme},
			BasePath: apiUrl.Path,
		}).PipelineService,
		context: ctx,
	}, nil
}

func (ps *PipelineService) PipelineIdForName(pipelineName string) (string, error) {
	pipelineResult, err := ps.ListPipelines(&pipeline_service.ListPipelinesParams{
		Filter:  byNameFilter(pipelineName),
		Context: ps.context,
	}, nil)
	if err != nil {
		return "", err
	}
	numPipelines := len(pipelineResult.Payload.Pipelines)
	if numPipelines != 1 {
		return "", fmt.Errorf("found %d pipelines, expected exactly one", numPipelines)
	}

	return pipelineResult.Payload.Pipelines[0].ID, nil
}

func (ps *PipelineService) PipelineVersionIdForName(versionName string, pipelineId string) (string, error) {
	pipelineVersionResult, err := ps.ListPipelineVersions(&pipeline_service.ListPipelineVersionsParams{
		Filter:        pipelineVersionByNameFilter(versionName),
		ResourceKeyID: &pipelineId,
		Context:       ps.context,
	}, nil)
	if err != nil {
		return "", err
	}
	numPipelineVersions := len(pipelineVersionResult.Payload.Versions)
	if numPipelineVersions != 1 {
		return "", fmt.Errorf("found %d pipeline versions, expected exactly one", numPipelineVersions)
	}

	return pipelineVersionResult.Payload.Versions[0].ID, nil
}

func pipelineUploadService(providerConfig config.KfpProviderConfig) (*pipeline_upload_service.Client, error) {
	apiUrl, err := url.Parse(providerConfig.Parameters.RestKfpApiUrl)
	if err != nil {
		return nil, err
	}

	return pipeline_upload_client.NewHTTPClientWithConfig(strfmt.NewFormats(), &pipeline_upload_client.TransportConfig{
		Host:     apiUrl.Host,
		Schemes:  []string{apiUrl.Scheme},
		BasePath: apiUrl.Path,
	}).PipelineUploadService, nil
}
