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
	"log"
	"os"
)

func main() {
	if err := RunProviderApp[VertexAiProviderConfig](VAIProvider{}); err != nil {
		log.Fatal(err)
	}
}

type VertexAiProviderConfig struct {
	PipelineBucket string `yaml:"pipelineBucket,omitempty"`
	Endpoint       string `yaml:"endpoint,omitempty"`
}

type VAIProvider struct {
}

func (vaip VAIProvider) client(providerConfig VertexAiProviderConfig, ctx context.Context) (*storage.Client, error) {
	var client *storage.Client
	var err error
	if providerConfig.Endpoint != "" {
		client, err = storage.NewClient(ctx, option.WithoutAuthentication(), option.WithEndpoint(providerConfig.Endpoint))
	} else {
		client, err = storage.NewClient(ctx)
	}

	return client, err
}

func (vaip VAIProvider) CreatePipeline(providerConfig VertexAiProviderConfig, pipelineDefinition PipelineDefinition, pipelineFile string, ctx context.Context) (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	return vaip.UpdatePipeline(providerConfig, pipelineDefinition, id.String(), pipelineFile, ctx)
}

func (vaip VAIProvider) UpdatePipeline(providerConfig VertexAiProviderConfig, pipelineDefinition PipelineDefinition, id string, pipelineFile string, ctx context.Context) (string, error) {
	client, err := vaip.client(providerConfig, ctx)
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

	fmt.Printf("Version %s created for pipeline %s\n", pipelineDefinition.Version, id)

	return id, nil
}

func (vaip VAIProvider) DeletePipeline(providerConfig VertexAiProviderConfig, id string, ctx context.Context) error {
	client, err := vaip.client(providerConfig, ctx)
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

func (vaip VAIProvider) CreateRunConfiguration(_ VertexAiProviderConfig, _ RunConfigurationDefinition, _ context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip VAIProvider) UpdateRunConfiguration(_ VertexAiProviderConfig, _ RunConfigurationDefinition, _ string, _ context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip VAIProvider) DeleteRunConfiguration(_ VertexAiProviderConfig, _ string, _ context.Context) error {
	return errors.New("not implemented")
}

func (vaip VAIProvider) CreateExperiment(_ VertexAiProviderConfig, _ ExperimentDefinition, _ context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip VAIProvider) UpdateExperiment(_ VertexAiProviderConfig, _ ExperimentDefinition, _ string, _ context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip VAIProvider) DeleteExperiment(_ VertexAiProviderConfig, _ string, _ context.Context) error {
	return errors.New("not implemented")
}
