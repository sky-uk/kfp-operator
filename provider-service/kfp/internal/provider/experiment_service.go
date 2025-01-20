package provider

import (
	"context"
	"fmt"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/util"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
)

type ExperimentService interface {
	CreateExperiment(
		experiment common.NamespacedName,
		description string,
	) (string, error)

	DeleteExperiment(id string) error

	ExperimentIdByName(experiment common.NamespacedName) (string, error)
}

type DefaultExperimentService struct {
	ctx    context.Context
	client client.ExperimentServiceClient
}

func NewExperimentService(
	ctx context.Context,
	providerConfig config.KfpProviderConfig,
) (*DefaultExperimentService, error) {
	// apiUrl, err := url.Parse(providerConfig.Parameters.RestKfpApiUrl)
	// if err != nil {
	// 	return nil, err
	// }

	// foo := go_client.NewExperimentServiceClient(conn)

	// TODO: GrpcKfpApi needs to instantiate in a way which shares out the
	// grpc connection so that the ExperimentService also has access.
	// event flow uses GrpcKfpApi, but not the experiment service - which is
	// why they are separate.
	return &DefaultExperimentService{
		// Client: experiment_client.NewHTTPClientWithConfig(
		// 	strfmt.NewFormats(), &experiment_client.TransportConfig{
		// 		Host:     apiUrl.Host,
		// 		Schemes:  []string{apiUrl.Scheme},
		// 		BasePath: apiUrl.Path,
		// 	},
		// ).ExperimentService,

		ctx: ctx,
	}, nil
}

func (es *DefaultExperimentService) CreateExperiment(
	experiment common.NamespacedName,
	description string,
) (string, error) {
	experimentName, err := util.ResourceNameFromNamespacedName(experiment)
	if err != nil {
		return "", err
	}

	result, err := es.client.CreateExperiment(
		es.ctx,
		&go_client.CreateExperimentRequest{
			Experiment: &go_client.Experiment{
				Name:        experimentName,
				Description: description,
			},
		},
	)
	if err != nil {
		return "", err
	}

	return result.Id, nil
}

func (es *DefaultExperimentService) DeleteExperiment(id string) error {
	_, err := es.client.DeleteExperiment(
		es.ctx,
		&go_client.DeleteExperimentRequest{
			Id: id,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (es *DefaultExperimentService) ExperimentIdByName(
	experiment common.NamespacedName,
) (string, error) {
	experimentName, err := util.ResourceNameFromNamespacedName(experiment)
	if err != nil {
		return "", err
	}

	experimentResult, err := es.client.ListExperiment(
		es.ctx,
		&go_client.ListExperimentsRequest{
			Filter: *byNameFilter(experimentName),
		},
		nil,
	)
	if err != nil {
		return "", err
	}

	numExperiments := len(experimentResult.Experiments)
	if numExperiments != 1 {
		return "", fmt.Errorf("found %d experiments, expected exactly one", numExperiments)
	}

	return experimentResult.Experiments[0].Id, nil
}
