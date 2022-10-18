package main

import (
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	. "github.com/sky-uk/kfp-operator/providers/base"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io"
	"os"
)

func main() {
	RunProviderApp[VertexAiProviderConfig](VAIProvider{})
}

type VertexAiProviderConfig struct {
	PipelineBucket string `yaml:"pipelineBucket,omitempty"`
	Endpoint       string `yaml:"endpoint,omitempty"`
}

type VAIProvider struct {
}

func (vaip VAIProvider) client(ctx context.Context, providerConfig VertexAiProviderConfig) (*storage.Client, error) {
	var client *storage.Client
	var err error
	if providerConfig.Endpoint != "" {
		client, err = storage.NewClient(ctx, option.WithoutAuthentication(), option.WithEndpoint(providerConfig.Endpoint))
	} else {
		client, err = storage.NewClient(ctx)
	}

	return client, err
}

func (vaip VAIProvider) CreatePipeline(ctx context.Context, providerConfig VertexAiProviderConfig, pipelineDefinition PipelineDefinition, pipelineFile string) (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	return vaip.UpdatePipeline(ctx, providerConfig, pipelineDefinition, id.String(), pipelineFile)
}

func (vaip VAIProvider) UpdatePipeline(ctx context.Context, providerConfig VertexAiProviderConfig, pipelineDefinition PipelineDefinition, id string, pipelineFile string) (string, error) {
	client, err := vaip.client(ctx, providerConfig)
	if err != nil {
		return id, err
	}

	reader, err := os.Open(pipelineFile)
	if err != nil {
		return id, err
	}

	writer := client.Bucket(providerConfig.PipelineBucket).Object(fmt.Sprintf("%s/%s", id, pipelineDefinition.Version)).NewWriter(ctx)
	_, err = io.Copy(writer, reader)
	if err != nil {
		return id, err
	}

	err = writer.Close()
	if err != nil {
		return id, err
	}

	err = reader.Close()
	if err != nil {
		return id, err
	}

	return id, nil
}

func (vaip VAIProvider) DeletePipeline(ctx context.Context, providerConfig VertexAiProviderConfig, id string) error {
	client, err := vaip.client(ctx, providerConfig)
	if err != nil {
		return err
	}

	query := &storage.Query{Prefix: fmt.Sprintf("%s/", id)}

	it := client.Bucket(providerConfig.PipelineBucket).Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		err = client.Bucket(providerConfig.PipelineBucket).Object(attrs.Name).Delete(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (vaip VAIProvider) CreateRunConfiguration(_ context.Context, _ VertexAiProviderConfig, _ RunConfigurationDefinition) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip VAIProvider) UpdateRunConfiguration(_ context.Context, _ VertexAiProviderConfig, _ RunConfigurationDefinition, _ string) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip VAIProvider) DeleteRunConfiguration(_ context.Context, _ VertexAiProviderConfig, _ string) error {
	return errors.New("not implemented")
}

func (vaip VAIProvider) CreateExperiment(_ context.Context, _ VertexAiProviderConfig, _ ExperimentDefinition) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip VAIProvider) UpdateExperiment(_ context.Context, _ VertexAiProviderConfig, _ ExperimentDefinition, _ string) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip VAIProvider) DeleteExperiment(_ context.Context, _ VertexAiProviderConfig, _ string) error {
	return errors.New("not implemented")
}
