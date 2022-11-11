package main

import (
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/scheduler/apiv1"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	. "github.com/sky-uk/kfp-operator/providers/base"
	"github.com/urfave/cli"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	schedulerpb "google.golang.org/genproto/googleapis/cloud/scheduler/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"io"
	"os"
)

var labels = struct {
	RunConfiguration string
	PipelineName     string
	PipelineVersion  string
}{
	RunConfiguration: "run-configuration",
	PipelineName:     "pipeline-name",
	PipelineVersion:  "pipeline-version",
}

func main() {
	app := NewProviderApp[VAIProviderConfig]()
	provider := VAIProvider{}
	app.Run(provider, cli.Command{
		Name: "vai-run",
		Subcommands: []cli.Command{
			{
				Name: "enqueue",
				Flags: []cli.Flag{cli.StringFlag{
					Name:     "run-intent",
					Required: true,
				}},
				Action: func(c *cli.Context) error {
					providerConfig, err := app.LoadProviderConfig(c)
					if err != nil {
						return err
					}
					runIntent, err := LoadJsonFromFile[RunIntent](c.String("run-intent"))
					if err != nil {
						return err
					}
					return provider.enqueueRun(app.Context, providerConfig, runIntent)
				},
			},
			{
				Name: "submit",
				Flags: []cli.Flag{cli.StringFlag{
					Name:     "run",
					Required: true,
				}},
				Action: func(c *cli.Context) error {
					providerConfig, err := app.LoadProviderConfig(c)
					vaiRun, err := LoadYamlFromFile[VAIRun](c.String("run"))
					if err != nil {
						return err
					}
					return provider.submitRun(app.Context, providerConfig, vaiRun)
				},
			},
		},
	})
}

type VAIProviderConfig struct {
	PipelineBucket  string `yaml:"pipelineBucket"`
	Endpoint        string `yaml:"endpoint"`
	RunIntentsTopic string `yaml:"runIntentsTopic"`
	Project         string `yaml:"project"`
	Location        string `yaml:"location"`
	RunsTopic       string `yaml:"runsTopic"`
}

type VAIProvider struct {
}

func (vaip VAIProvider) parent(providerConfig VAIProviderConfig) string {
	//TODO load from default?
	return fmt.Sprintf(`projects/%s/locations/%s`, providerConfig.Project, providerConfig.Location)
}

func (vaip VAIProvider) jobName(providerConfig VAIProviderConfig, name string) string {
	return fmt.Sprintf("%s/jobs/%s", vaip.parent(providerConfig), name)
}

func (vaip VAIProvider) topic(providerConfig VAIProviderConfig, topicName string) string {
	return fmt.Sprintf("projects/%s/topics/%s", providerConfig.Project, topicName)
}

func (vaip VAIProvider) pipelineStorageObject(pipelineName string, pipelineVersion string) string {
	return fmt.Sprintf("%s/%s", pipelineName, pipelineVersion)
}

func (vaip VAIProvider) pipelineUri(providerConfig VAIProviderConfig, pipelineName string, pipelineVersion string) string {
	return fmt.Sprintf("gs://%s/%s", providerConfig.PipelineBucket, vaip.pipelineStorageObject(pipelineName, pipelineVersion))
}

func (vaip VAIProvider) gcsClient(ctx context.Context, providerConfig VAIProviderConfig) (*storage.Client, error) {
	var client *storage.Client
	var err error
	if providerConfig.Endpoint != "" {
		client, err = storage.NewClient(ctx, option.WithoutAuthentication(), option.WithEndpoint(providerConfig.Endpoint))
	} else {
		client, err = storage.NewClient(ctx)
	}

	return client, err
}

func (vaip VAIProvider) CreatePipeline(ctx context.Context, providerConfig VAIProviderConfig, pipelineDefinition PipelineDefinition, pipelineFile string) (string, error) {
	if _, err := vaip.UpdatePipeline(ctx, providerConfig, pipelineDefinition, pipelineDefinition.Name, pipelineFile); err != nil {
		return "", err
	}

	return pipelineDefinition.Name, nil
}

func (vaip VAIProvider) UpdatePipeline(ctx context.Context, providerConfig VAIProviderConfig, pipelineDefinition PipelineDefinition, id string, pipelineFile string) (string, error) {
	client, err := vaip.gcsClient(ctx, providerConfig)
	if err != nil {
		return id, err
	}

	reader, err := os.Open(pipelineFile)
	if err != nil {
		return id, err
	}

	writer := client.Bucket(providerConfig.PipelineBucket).Object(vaip.pipelineStorageObject(id, pipelineDefinition.Version)).NewWriter(ctx)
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

func (vaip VAIProvider) DeletePipeline(ctx context.Context, providerConfig VAIProviderConfig, id string) error {
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

type VAIRun struct {
	RunId             string            `json:"runId"`
	Labels            map[string]string `json:"labels,omitempty"`
	PipelineUri       string            `json:"pipelineUri"`
	RuntimeParameters map[string]string `json:"runtimeParameters,omitempty"`
}

type RunIntent struct {
	RunConfigurationName string            `json:"runConfigurationName,omitempty"`
	PipelineName         string            `json:"pipelineName"`
	PipelineVersion      string            `json:"pipelineVersion"`
	RuntimeParameters    map[string]string `json:"runtimeParameters,omitempty"`
}

func (vaip VAIProvider) createSchedulerJobPb(providerConfig VAIProviderConfig, runConfigurationDefinition RunConfigurationDefinition) (*schedulerpb.Job, error) {
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

func (vaip VAIProvider) CreateRunConfiguration(ctx context.Context, providerConfig VAIProviderConfig, runConfigurationDefinition RunConfigurationDefinition) (string, error) {
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

func (vaip VAIProvider) UpdateRunConfiguration(ctx context.Context, providerConfig VAIProviderConfig, runConfigurationDefinition RunConfigurationDefinition, providerId string) (string, error) {
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

func (vaip VAIProvider) DeleteRunConfiguration(ctx context.Context, _ VAIProviderConfig, providerId string) error {
	client, err := scheduler.NewCloudSchedulerClient(ctx)
	if err != nil {
		return err
	}

	return client.DeleteJob(ctx, &schedulerpb.DeleteJobRequest{
		Name: providerId,
	})
}

func (vaip VAIProvider) CreateExperiment(_ context.Context, _ VAIProviderConfig, _ ExperimentDefinition) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip VAIProvider) UpdateExperiment(_ context.Context, _ VAIProviderConfig, _ ExperimentDefinition, _ string) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip VAIProvider) DeleteExperiment(_ context.Context, _ VAIProviderConfig, _ string) error {
	return errors.New("not implemented")
}

func (vaip VAIProvider) enqueueRun(ctx context.Context, providerConfig VAIProviderConfig, runIntent RunIntent) error {
	pubsubClient, err := pubsub.NewClient(ctx, providerConfig.Project)
	if err != nil {
		return err
	}

	topic := pubsubClient.Topic(providerConfig.RunsTopic)
	defer topic.Stop()

	vaiRun := VAIRun{
		RunId:       fmt.Sprintf("rc-%s-%s", runIntent.RunConfigurationName, uuid.New().String()),
		PipelineUri: vaip.pipelineUri(providerConfig, runIntent.PipelineName, runIntent.PipelineVersion),
		Labels: map[string]string{
			labels.RunConfiguration: runIntent.RunConfigurationName,
			labels.PipelineName:     runIntent.PipelineName,
			labels.PipelineVersion:  runIntent.PipelineVersion,
		},
		RuntimeParameters: runIntent.RuntimeParameters,
	}

	payload, err := json.Marshal(vaiRun)
	if err != nil {
		return err
	}

	_, err = topic.Publish(ctx, &pubsub.Message{Data: payload}).Get(ctx)
	return err
}

func (vaip VAIProvider) submitRun(ctx context.Context, providerConfig VAIProviderConfig, vaiRun VAIRun) error {
	return nil
}
