package provider

import (
	"context"
	"fmt"

	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/util"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	kfpUtil "github.com/sky-uk/kfp-operator/provider-service/kfp/internal/util"
	"google.golang.org/grpc"
)

type ExperimentService interface {
	CreateExperiment(
		ctx context.Context,
		experiment common.NamespacedName,
		description string,
	) (string, error)

	DeleteExperiment(ctx context.Context, id string) error

	ExperimentIdByDisplayName(ctx context.Context, experiment common.NamespacedName) (string, error)
}

type DefaultExperimentService struct {
	client client.ExperimentServiceClient
}

func NewExperimentService(
	conn *grpc.ClientConn,
) (ExperimentService, error) {
	if conn == nil {
		return nil, fmt.Errorf(
			"no gRPC connection was provided to start experiment service",
		)
	}

	return &DefaultExperimentService{
		client: go_client.NewExperimentServiceClient(conn),
	}, nil
}

// CreateExperiment creates an experiment and returns an experiment id
func (es *DefaultExperimentService) CreateExperiment(
	ctx context.Context,
	experiment common.NamespacedName,
	description string,
) (string, error) {
	experimentName, err := util.ResourceNameFromNamespacedName(experiment)
	if err != nil {
		return "", err
	}

	result, err := es.client.CreateExperiment(
		ctx,
		&go_client.CreateExperimentRequest{
			Experiment: &go_client.Experiment{
				DisplayName: experimentName,
				Description: description,
			},
		},
	)
	if err != nil {
		return "", err
	}

	return result.ExperimentId, nil
}

// Delete Experiment deletes an experiment by experiment id
func (es *DefaultExperimentService) DeleteExperiment(ctx context.Context, id string) error {
	_, err := es.client.DeleteExperiment(
		ctx,
		&go_client.DeleteExperimentRequest{ExperimentId: id},
	)
	if err != nil {
		return err
	}

	return nil
}

// ExperimentIdByDisplayName gets the experiment id corresponding to the experiment name.
// Expects to find exactly one such experiment.
func (es *DefaultExperimentService) ExperimentIdByDisplayName(
	ctx context.Context,
	experiment common.NamespacedName,
) (string, error) {
	experimentName, err := util.ResourceNameFromNamespacedName(experiment)
	if err != nil {
		return "", err
	}

	experimentResult, err := es.client.ListExperiments(
		ctx,
		&go_client.ListExperimentsRequest{
			Filter: kfpUtil.ByDisplayNameFilter(experimentName),
		},
	)
	if err != nil {
		return "", err
	}

	experimentCount := len(experimentResult.Experiments)
	if experimentCount != 1 {
		return "", fmt.Errorf("found %d experiments, expected exactly one", experimentCount)
	}

	return experimentResult.Experiments[0].ExperimentId, nil
}
