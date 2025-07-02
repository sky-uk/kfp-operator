package provider

import (
	"context"
	"errors"

	"github.com/sky-uk/kfp-operator/pkg/providers/base"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type KfpProvider struct {
	experimentService ExperimentService
}

func NewKfpProvider(config config.Config) (*KfpProvider, error) {
	conn, err := grpc.NewClient(
		config.Parameters.GrpcKfpApiAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	experimentService, err := NewExperimentService(conn)

	return &KfpProvider{
		experimentService: experimentService,
	}, nil
}

var _ resource.Provider = &KfpProvider{}

func (*KfpProvider) CreatePipeline(
	ctx context.Context,
	ppd resource.PipelineDefinitionWrapper,
) (string, error) {
	return "", errors.New("not implemented")
}

func (*KfpProvider) UpdatePipeline(
	ctx context.Context,
	ppd resource.PipelineDefinitionWrapper,
	id string,
) (string, error) {
	return "", errors.New("not implemented")
}

func (*KfpProvider) DeletePipeline(
	ctx context.Context,
	id string,
) error {
	return errors.New("not implemented")
}

func (*KfpProvider) CreateRun(
	ctx context.Context,
	rd base.RunDefinition,
) (string, error) {
	return "", errors.New("not implemented")
}

func (*KfpProvider) DeleteRun(
	ctx context.Context,
	id string,
) error {
	return errors.New("not implemented")
}

func (*KfpProvider) CreateRunSchedule(
	ctx context.Context,
	rsd base.RunScheduleDefinition,
) (string, error) {
	return "", errors.New("not implemented")
}

func (*KfpProvider) UpdateRunSchedule(
	ctx context.Context,
	rsd base.RunScheduleDefinition,
	id string,
) (string, error) {
	return "", errors.New("not implemented")
}

func (*KfpProvider) DeleteRunSchedule(
	ctx context.Context,
	id string,
) error {
	return errors.New("not implemented")
}

func (p *KfpProvider) CreateExperiment(
	ctx context.Context,
	ed base.ExperimentDefinition,
) (string, error) {
	expId, err := p.experimentService.CreateExperiment(
		ctx,
		ed.Name,
		ed.Description,
	)
	if err != nil {
		return "", err
	}

	return expId, nil
}

func (p *KfpProvider) UpdateExperiment(
	ctx context.Context,
	ed base.ExperimentDefinition,
	id string,
) (string, error) {
	if err := p.DeleteExperiment(ctx, id); err != nil {
		return id, err
	}

	return p.CreateExperiment(ctx, ed)
}

func (p *KfpProvider) DeleteExperiment(
	ctx context.Context,
	id string,
) error {
	return p.experimentService.DeleteExperiment(ctx, id)
}
