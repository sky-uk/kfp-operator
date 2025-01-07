package provider

import (
	"context"
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_client/experiment_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/job_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/job_client/job_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/run_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/run_client/run_service"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"net/url"
)

type ExperimentService struct {
	*experiment_service.Client
	context context.Context
}

func NewExperimentService(ctx context.Context, providerConfig config.KfpProviderConfig) (*ExperimentService, error) {
	apiUrl, err := url.Parse(providerConfig.Parameters.RestKfpApiUrl)
	if err != nil {
		return nil, err
	}

	return &ExperimentService{Client: experiment_client.NewHTTPClientWithConfig(strfmt.NewFormats(), &experiment_client.TransportConfig{
		Host:     apiUrl.Host,
		Schemes:  []string{apiUrl.Scheme},
		BasePath: apiUrl.Path,
	}).ExperimentService,
		context: ctx,
	}, nil
}

func (es *ExperimentService) ExperimentIdByName(experimentNamespacedName common.NamespacedName) (string, error) {
	experimentName, err := ResourceNameFromNamespacedName(experimentNamespacedName)
	if err != nil {
		return "", err
	}

	experimentResult, err := es.ListExperiment(&experiment_service.ListExperimentParams{
		Filter:  byNameFilter(experimentName),
		Context: es.context,
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

func runService(providerConfig config.KfpProviderConfig) (*run_service.Client, error) {
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

func jobService(providerConfig config.KfpProviderConfig) (*job_service.Client, error) {
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
