package kfp

import (
	"context"
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
	"github.com/kubeflow/pipelines/backend/api/go_http_client/run_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/run_client/run_service"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/argo/providers/base"
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
}

func NewPipelineService(providerConfig KfpProviderConfig) (*PipelineService, error) {
	apiUrl, err := url.Parse(providerConfig.Parameters.RestKfpApiUrl)
	if err != nil {
		return nil, err
	}

	return &PipelineService{
		pipeline_client.NewHTTPClientWithConfig(strfmt.NewFormats(), &pipeline_client.TransportConfig{
			Host:     apiUrl.Host,
			Schemes:  []string{apiUrl.Scheme},
			BasePath: apiUrl.Path,
		}).PipelineService}, nil
}

func (ps *PipelineService) PipelineIdForName(ctx context.Context, pipelineName string) (string, error) {
	pipelineResult, err := ps.ListPipelines(&pipeline_service.ListPipelinesParams{
		Filter:  byNameFilter(pipelineName),
		Context: ctx,
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

func (ps *PipelineService) PipelineVersionIdForName(ctx context.Context, versionName string, pipelineId string) (string, error) {
	pipelineVersionResult, err := ps.ListPipelineVersions(&pipeline_service.ListPipelineVersionsParams{
		Filter:        pipelineVersionByNameFilter(versionName),
		ResourceKeyID: &pipelineId,
		Context:       ctx,
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

func pipelineUploadService(providerConfig KfpProviderConfig) (*pipeline_upload_service.Client, error) {
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

type ExperimentService struct {
	*experiment_service.Client
}

func NewExperimentService(providerConfig KfpProviderConfig) (*ExperimentService, error) {
	apiUrl, err := url.Parse(providerConfig.Parameters.RestKfpApiUrl)
	if err != nil {
		return nil, err
	}

	return &ExperimentService{experiment_client.NewHTTPClientWithConfig(strfmt.NewFormats(), &experiment_client.TransportConfig{
		Host:     apiUrl.Host,
		Schemes:  []string{apiUrl.Scheme},
		BasePath: apiUrl.Path,
	}).ExperimentService}, nil
}

func (es *ExperimentService) ExperimentIdByName(ctx context.Context, experimentNamespacedName common.NamespacedName) (string, error) {
	experimentName, err := ResourceNameFromNamespacedName(experimentNamespacedName)
	if err != nil {
		return "", err
	}

	experimentResult, err := es.ListExperiment(&experiment_service.ListExperimentParams{
		Filter:  byNameFilter(experimentName),
		Context: ctx,
	}, nil)
	if err != nil {
		return "", err
	}

	numExperiments := len(experimentResult.Payload.Experiments)
	if numExperiments != 1 {
		return "", fmt.Errorf("found %d experiments, expected exactly one", numExperiments)
	}

	return experimentResult.Payload.Experiments[0].ID, nil
}

func runService(providerConfig KfpProviderConfig) (*run_service.Client, error) {
	apiUrl, err := url.Parse(providerConfig.Parameters.RestKfpApiUrl)
	if err != nil {
		return nil, err
	}

	return run_client.NewHTTPClientWithConfig(strfmt.NewFormats(), &run_client.TransportConfig{
		Host:     apiUrl.Host,
		Schemes:  []string{apiUrl.Scheme},
		BasePath: apiUrl.Path,
	}).RunService, nil
}

func jobService(providerConfig KfpProviderConfig) (*job_service.Client, error) {
	apiUrl, err := url.Parse(providerConfig.Parameters.RestKfpApiUrl)
	if err != nil {
		return nil, err
	}

	return job_client.NewHTTPClientWithConfig(strfmt.NewFormats(), &job_client.TransportConfig{
		Host:     apiUrl.Host,
		Schemes:  []string{apiUrl.Scheme},
		BasePath: apiUrl.Path,
	}).JobService, nil
}
