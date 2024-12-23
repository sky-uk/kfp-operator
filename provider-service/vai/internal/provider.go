package internal

import (
	"cloud.google.com/go/storage"
	"context"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resources"
	"google.golang.org/api/option"
	"io"
)

type VAIProvider struct {
	ctx       context.Context
	config    VAIProviderConfig
	gcsClient *storage.Client
}

func NewProvider(ctx context.Context, config VAIProviderConfig) (*VAIProvider, error) {
	gcsClient, err := gcsClient(ctx, config)
	if err != nil {
		return nil, err
	}

	return &VAIProvider{
		ctx:       ctx,
		config:    config,
		gcsClient: gcsClient,
	}, nil
}

func (vaip *VAIProvider) CreatePipeline(pd resources.PipelineDefinition) (string, error) {
	pipelineId, err := vaip.UpdatePipeline(pd, "")
	if err != nil {
		return "", err
	}
	return pipelineId, nil
}

func (vaip *VAIProvider) UpdatePipeline(pd resources.PipelineDefinition, id string) (string, error) {
	pipelineId, err := pd.Name.String()
	if err != nil {
		return "", err
	}

	storageObject, err := vaip.config.pipelineStorageObject(pd.Name, pd.Version)
	if err != nil {
		return pipelineId, err
	}

	writer := vaip.gcsClient.Bucket(vaip.config.Parameters.PipelineBucket).Object(storageObject).NewWriter(vaip.ctx)

	_, err = io.Writer(writer).Write(pd.Manifest)
	if err != nil {
		return "", err
	}

	if err = writer.Close(); err != nil {
		return "pipelineId", err
	}

	return pipelineId, nil
}

func (vaip *VAIProvider) DeletePipeline(id string) error {
	return nil
}

func (vaip *VAIProvider) CreateRun(rcd resources.RunDefinition) (string, error) {
	return "", nil
}

func (vaip *VAIProvider) DeleteRun(id string) error {
	return nil
}

func (vaip *VAIProvider) CreateRunSchedule(rsd resources.RunScheduleDefinition) (string, error) {
	return "", nil
}

func (vaip *VAIProvider) UpdateRunSchedule(rsd resources.RunScheduleDefinition, id string) (string, error) {
	return "", nil
}

func (vaip *VAIProvider) DeleteRunSchedule(id string) error {
	return nil
}

func (vaip *VAIProvider) CreateExperiment(ed resources.ExperimentDefinition) (string, error) {
	return "", nil
}

func (vaip *VAIProvider) UpdateExperiment(ed resources.ExperimentDefinition, id string) (string, error) {
	return "", nil
}

func (vaip *VAIProvider) DeleteExperiment(id string) error {
	return nil
}

func gcsClient(ctx context.Context, providerConfig VAIProviderConfig) (*storage.Client, error) {
	var client *storage.Client
	var err error
	if providerConfig.Parameters.GcsEndpoint != "" {
		client, err = storage.NewClient(ctx, option.WithoutAuthentication(), option.WithEndpoint(providerConfig.Parameters.GcsEndpoint))
	} else {
		client, err = storage.NewClient(ctx)
	}

	return client, err
}
