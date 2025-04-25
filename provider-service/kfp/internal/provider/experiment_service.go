package provider

import (
	"context"
	"fmt"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/util"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	kfpUtil "github.com/sky-uk/kfp-operator/provider-service/kfp/internal/provider/util"
	"google.golang.org/grpc"
)

type ExperimentService interface {
	CreateExperiment(
		ctx context.Context,
		experiment common.NamespacedName,
		description string,
	) (string, error)

	DeleteExperiment(ctx context.Context, id string) error

	ExperimentIdByName(ctx context.Context, experiment common.NamespacedName) (string, error)
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

// Delete Experiment deletes an experiment by experiment id
func (es *DefaultExperimentService) DeleteExperiment(ctx context.Context, id string) error {
	_, err := es.client.DeleteExperiment(
		ctx,
		&go_client.DeleteExperimentRequest{Id: id},
	)
	if err != nil {
		return err
	}

	return nil
}

// ExperimentIdByName gets the experiment id corresponding to the experiment name.
// Expects to find exactly one such experiment.
func (es *DefaultExperimentService) ExperimentIdByName(
	ctx context.Context,
	experiment common.NamespacedName,
) (string, error) {
	experimentName, err := util.ResourceNameFromNamespacedName(experiment)
	if err != nil {
		return "", err
	}

	experimentResult, err := es.client.ListExperiment(
		ctx,
		&go_client.ListExperimentsRequest{
			Filter: *kfpUtil.ByNameFilter(experimentName),
		},
	)
	if err != nil {
		return "", err
	}

	experimentCount := len(experimentResult.Experiments)
	if experimentCount != 1 {
		return "", fmt.Errorf("found %d experiments, expected exactly one", experimentCount)
	}

	return experimentResult.Experiments[0].Id, nil
}
