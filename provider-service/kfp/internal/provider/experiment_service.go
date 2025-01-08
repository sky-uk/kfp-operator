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

type ExperimentService struct {
	context context.Context
	client.ExperimentServiceClient
}

func NewExperimentService(ctx context.Context, providerConfig config.KfpProviderConfig) (*ExperimentService, error) {
	// apiUrl, err := url.Parse(providerConfig.Parameters.RestKfpApiUrl)
	// if err != nil {
	// 	return nil, err
	// }

	// foo := go_client.NewExperimentServiceClient(conn)

	// TODO: GrpcKfpApi needs to instantiate in a way which shares out the
	// grpc connection so that the ExperimentService also has access.
	// event flow uses GrpcKfpApi, but not the experiment service - which is
	// why they are separate.
	return &ExperimentService{
		// Client: experiment_client.NewHTTPClientWithConfig(
		// 	strfmt.NewFormats(), &experiment_client.TransportConfig{
		// 		Host:     apiUrl.Host,
		// 		Schemes:  []string{apiUrl.Scheme},
		// 		BasePath: apiUrl.Path,
		// 	},
		// ).ExperimentService,

		context: ctx,
	}, nil
}

func (es *ExperimentService) ExperimentIdByName(experimentNamespacedName common.NamespacedName) (string, error) {
	experimentName, err := util.ResourceNameFromNamespacedName(experimentNamespacedName)
	if err != nil {
		return "", err
	}

	experimentResult, err := es.ListExperiment(
		es.context,
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
