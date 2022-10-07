package main

import (
	"cloud.google.com/go/storage"
	"context"
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
	err := RunProviderApp[VertexAiProviderConfig](VAIProvider{})

	if err != nil {
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

func (vaip VAIProvider) CreatePipeline(_ VertexAiProviderConfig, _ PipelineConfig, _ string, _ context.Context) (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	return id.String(), nil
}

func (vaip VAIProvider) UpdatePipeline(providerConfig VertexAiProviderConfig, pipelineConfig PipelineConfig, id string, pipelineFile string, ctx context.Context) (string, error) {
	client, err := vaip.client(providerConfig, ctx)
	if err != nil {
		log.Fatal(err)
	}

	reader, err := os.Open(pipelineFile)
	if err != nil {
		return "", err
	}

	writer := client.Bucket(providerConfig.PipelineBucket).Object(fmt.Sprintf("%s/%s", id, pipelineConfig.Version)).NewWriter(ctx)
	_, err = io.Copy(writer, reader)
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	err = reader.Close()
	if err != nil {
		return "", err
	}

	fmt.Printf("Version %s created for pipeline %s\n", pipelineConfig.Version, id)

	return pipelineConfig.Version, nil
}

func (vaip VAIProvider) DeletePipeline(providerConfig VertexAiProviderConfig, _ PipelineConfig, id string, ctx context.Context) error {
	client, err := vaip.client(providerConfig, ctx)
	if err != nil {
		log.Fatal(err)
	}

	query := &storage.Query{Prefix: id}

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
