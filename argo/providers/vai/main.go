package main

import (
	"cloud.google.com/go/scheduler/apiv1"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	. "github.com/sky-uk/kfp-operator/providers/base"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	schedulerpb "google.golang.org/genproto/googleapis/cloud/scheduler/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"io"
	"os"
)

func main() {
	RunProviderApp[VertexAiProviderConfig](VAIProvider{})
}

type VertexAiProviderConfig struct {
	PipelineBucket  string `yaml:"pipelineBucket,omitempty"`
	Endpoint        string `yaml:"endpoint,omitempty"`
	RunIntentsTopic string `yaml:"runIntentsTopic,omitempty"`
	Project         string `yaml:"project,omitempty"`
	Location        string `yaml:"location,omitempty"`
}

type VAIProvider struct {
}

func (vaip VAIProvider) parent(providerConfig VertexAiProviderConfig) string {
	//TODO load from default?
	return fmt.Sprintf(`projects/%s/locations/%s`, providerConfig.Project, providerConfig.Location)
}

func (vaip VAIProvider) jobName(providerConfig VertexAiProviderConfig, name string) string {
	return fmt.Sprintf("%s/jobs/%s", vaip.parent(providerConfig), name)
}

func (vaip VAIProvider) topic(providerConfig VertexAiProviderConfig, topicName string) string {
	return fmt.Sprintf("projects/%s/topics/%s", providerConfig.Project, topicName)
}

func (vaip VAIProvider) gcsClient(ctx context.Context, providerConfig VertexAiProviderConfig) (*storage.Client, error) {
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

	if _, err := vaip.UpdatePipeline(ctx, providerConfig, pipelineDefinition, id.String(), pipelineFile); err != nil {
		return "", err
	}

	return id.String(), nil
}

func (vaip VAIProvider) UpdatePipeline(ctx context.Context, providerConfig VertexAiProviderConfig, pipelineDefinition PipelineDefinition, id string, pipelineFile string) (string, error) {
	client, err := vaip.gcsClient(ctx, providerConfig)
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
	client, err := vaip.gcsClient(ctx, providerConfig)
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

type RunIntent struct {
	RunConfigurationName string            `json:"runConfigurationName,omitempty"`
	PipelineName         string            `json:"pipelineName,omitempty"`
	PipelineVersion      string            `json:"pipelineVersion,omitempty"`
	RuntimeParameters    map[string]string `json:"runtimeParameters,omitempty"`
}

func (vaip VAIProvider) createSchedulerJobPb(providerConfig VertexAiProviderConfig, runConfigurationDefinition RunConfigurationDefinition) (*schedulerpb.Job, error) {
	fmt.Println(runConfigurationDefinition.Schedule)
	schedule, err := ParseCron(runConfigurationDefinition.Schedule)
	if err != nil {
		return nil, err
	}

	runIntent := RunIntent{
		RunConfigurationName: runConfigurationDefinition.Name,
		PipelineName:         runConfigurationDefinition.PipelineName,
		PipelineVersion:      runConfigurationDefinition.PipelineVersion,
		RuntimeParameters:    map[string]string{}, // See https://github.com/sky-uk/kfp-operator/issues/175
	}

	data, err := json.Marshal(runIntent)
	if err != nil {
		return nil, err
	}

	return &schedulerpb.Job{
		Name:     vaip.jobName(providerConfig, fmt.Sprintf("rc-%s", runConfigurationDefinition.Name)),
		Schedule: schedule.PrintStandard(),
		Target: &schedulerpb.Job_PubsubTarget{
			PubsubTarget: &schedulerpb.PubsubTarget{
				TopicName: vaip.topic(providerConfig, providerConfig.RunIntentsTopic),
				Data:      data,
			},
		},
	}, nil
}

func (vaip VAIProvider) CreateRunConfiguration(ctx context.Context, providerConfig VertexAiProviderConfig, runConfigurationDefinition RunConfigurationDefinition) (string, error) {
	client, err := scheduler.NewCloudSchedulerClient(ctx)
	if err != nil {
		return "", err
	}

	jobPb, err := vaip.createSchedulerJobPb(providerConfig, runConfigurationDefinition)
	if err != nil {
		return "", err
	}

	job, err := client.CreateJob(ctx, &schedulerpb.CreateJobRequest{
		Parent: vaip.parent(providerConfig),
		Job:    jobPb,
	})

	if err != nil {
		return "", err
	}

	return job.Name, nil
}

func (vaip VAIProvider) UpdateRunConfiguration(ctx context.Context, providerConfig VertexAiProviderConfig, runConfigurationDefinition RunConfigurationDefinition, providerId string) (string, error) {
	client, err := scheduler.NewCloudSchedulerClient(ctx)
	if err != nil {
		return "", err
	}

	jobPb, err := vaip.createSchedulerJobPb(providerConfig, runConfigurationDefinition)
	if err != nil {
		return "", err
	}

	jobPb.Name = providerId

	job, err := client.UpdateJob(ctx, &schedulerpb.UpdateJobRequest{
		Job: jobPb,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				"schedule",
				"pubsub_target.data",
			},
		},
	})

	if err != nil {
		return "", err
	}

	return job.Name, nil
}

func (vaip VAIProvider) DeleteRunConfiguration(ctx context.Context, _ VertexAiProviderConfig, providerId string) error {
	client, err := scheduler.NewCloudSchedulerClient(ctx)
	if err != nil {
		return err
	}

	return client.DeleteJob(ctx, &schedulerpb.DeleteJobRequest{
		Name: providerId,
	})
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
