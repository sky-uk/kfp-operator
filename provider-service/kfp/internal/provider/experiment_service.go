package provider

import (
	"context"
	"fmt"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/util"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	"google.golang.org/grpc"
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
	conn *grpc.ClientConn,
) (ExperimentService, error) {
	if conn == nil {
		return nil, fmt.Errorf(
			"no gRPC connection was provided to start experiment service",
		)
	}

	return &DefaultExperimentService{
		ctx:    ctx,
		client: go_client.NewExperimentServiceClient(conn),
	}, nil
}

// CreateExperiment creates an experiment and returns an experiment id
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

// Delete Experiment deletes an experiment by experiment id
func (es *DefaultExperimentService) DeleteExperiment(id string) error {
	_, err := es.client.DeleteExperiment(
		es.ctx,
		&go_client.DeleteExperimentRequest{Id: id},
	)
	if err != nil {
		return err
	}

	return nil
}

func (es *DefaultExperimentService) ExperimentIdByName(
	experiment common.NamespacedName,
) (string, error) {
	logger := common.LoggerFromContext(es.ctx)
	experimentName, err := util.ResourceNameFromNamespacedName(experiment)
	if err != nil {
		return "", err
	}

	logger.Info("fetching experiments", "experiment name", experiment)

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
	if experimentResult == nil {
		logger.Error(
			fmt.Errorf("Did not receive a ListExperimentsResponse"),
			"no ListExperimentsResponse",
		)
		return "", fmt.Errorf("Did not receive a ListExperimentsResponse")
	}

	numExperiments := len(experimentResult.Experiments)
	if numExperiments != 1 {
		return "", fmt.Errorf("found %d experiments, expected exactly one", numExperiments)
	}

	return experimentResult.Experiments[0].Id, nil
}
