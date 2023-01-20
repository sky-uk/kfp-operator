package kfp

import (
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_client/experiment_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/job_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/job_client/job_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_client/pipeline_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_client/pipeline_upload_service"
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

func pipelineUploadService(providerConfig KfpProviderConfig) (*pipeline_upload_service.Client, error) {
	apiUrl, err := url.Parse(providerConfig.RestKfpApiUrl)
	if err != nil {
		return nil, err
	}

	return pipeline_upload_client.NewHTTPClientWithConfig(strfmt.NewFormats(), &pipeline_upload_client.TransportConfig{
		Host:     apiUrl.Host,
		Schemes:  []string{apiUrl.Scheme},
		BasePath: apiUrl.Path,
	}).PipelineUploadService, nil
}

func pipelineService(providerConfig KfpProviderConfig) (*pipeline_service.Client, error) {
	apiUrl, err := url.Parse(providerConfig.RestKfpApiUrl)
	if err != nil {
		return nil, err
	}

	return pipeline_client.NewHTTPClientWithConfig(strfmt.NewFormats(), &pipeline_client.TransportConfig{
		Host:     apiUrl.Host,
		Schemes:  []string{apiUrl.Scheme},
		BasePath: apiUrl.Path,
	}).PipelineService, nil
}

func experimentService(providerConfig KfpProviderConfig) (*experiment_service.Client, error) {
	apiUrl, err := url.Parse(providerConfig.RestKfpApiUrl)
	if err != nil {
		return nil, err
	}

	return experiment_client.NewHTTPClientWithConfig(strfmt.NewFormats(), &experiment_client.TransportConfig{
		Host:     apiUrl.Host,
		Schemes:  []string{apiUrl.Scheme},
		BasePath: apiUrl.Path,
	}).ExperimentService, nil
}

func jobService(providerConfig KfpProviderConfig) (*job_service.Client, error) {
	apiUrl, err := url.Parse(providerConfig.RestKfpApiUrl)
	if err != nil {
		return nil, err
	}

	return job_client.NewHTTPClientWithConfig(strfmt.NewFormats(), &job_client.TransportConfig{
		Host:     apiUrl.Host,
		Schemes:  []string{apiUrl.Scheme},
		BasePath: apiUrl.Path,
	}).JobService, nil
}
