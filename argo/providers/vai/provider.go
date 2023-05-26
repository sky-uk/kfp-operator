package vai

import (
	"bytes"
	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/scheduler/apiv1"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/argo/providers/base/generic"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	aiplatformpb "google.golang.org/genproto/googleapis/cloud/aiplatform/v1"
	schedulerpb "google.golang.org/genproto/googleapis/cloud/scheduler/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"io"
	"os"
	"regexp"
	"strings"
)

var labels = struct {
	RunConfiguration string
	PipelineName     string
	PipelineVersion  string
	Namespace        string
}{
	RunConfiguration: "run-configuration",
	PipelineName:     "pipeline-name",
	PipelineVersion:  "pipeline-version",
	Namespace:        "namespace",
}

type VAIRun struct {
	RunId             string            `json:"runId"`
	Labels            map[string]string `json:"labels,omitempty"`
	PipelineUri       string            `json:"pipelineUri"`
	RuntimeParameters map[string]string `json:"runtimeParameters,omitempty"`
}

type RunIntent struct {
	RunConfigurationName common.NamespacedName `json:"runConfigurationName,omitempty"`
	RunName              common.NamespacedName `json:"runName,omitempty"`
	PipelineName         string                `json:"pipelineName"`
	PipelineVersion      string                `json:"pipelineVersion"`
	RuntimeParameters    map[string]string     `json:"runtimeParameters,omitempty"`
}

type VAIProviderConfig struct {
	VaiProject                  string `yaml:"vaiProject"`
	VaiLocation                 string `yaml:"vaiLocation"`
	VaiJobServiceAccount        string `yaml:"vaiJobServiceAccount"`
	GcsEndpoint                 string `yaml:"gcsEndpoint"`
	PipelineBucket              string `yaml:"pipelineBucket"`
	RunIntentsTopic             string `yaml:"runIntentsTopic"`
	RunsTopic                   string `yaml:"runsTopic"`
	EventsourceRunsSubscription string `yaml:"eventsourceRunsSubscription"`
}

func (vaipc VAIProviderConfig) vaiEndpoint() string {
	return fmt.Sprintf("%s-aiplatform.googleapis.com:443", vaipc.VaiLocation)
}

func (vaipc VAIProviderConfig) parent() string {
	return fmt.Sprintf(`projects/%s/locations/%s`, vaipc.VaiProject, vaipc.VaiLocation)
}

func (vaipc VAIProviderConfig) pipelineJobName(name string) string {
	return fmt.Sprintf("%s/pipelineJobs/%s", vaipc.parent(), name)
}

func (vaipc VAIProviderConfig) schedulerJobName(name string) string {
	return fmt.Sprintf("%s/jobs/%s", vaipc.parent(), name)
}

func (vaipc VAIProviderConfig) runIntentsTopicFullName() string {
	return vaipc.topicFullName(vaipc.RunIntentsTopic)
}

func (vaipc VAIProviderConfig) topicFullName(topicName string) string {
	return fmt.Sprintf("projects/%s/topics/%s", vaipc.VaiProject, topicName)
}

func (vaipc VAIProviderConfig) pipelineStorageObject(pipelineName string, pipelineVersion string) string {
	return fmt.Sprintf("%s/%s", pipelineName, pipelineVersion)
}

func (vaipc VAIProviderConfig) gcsUri(bucket string, pathSegments ...string) string {
	return fmt.Sprintf("gs://%s/%s", bucket, strings.Join(pathSegments, "/"))
}

func (vaipc VAIProviderConfig) pipelineUri(pipelineName string, pipelineVersion string) string {
	return vaipc.gcsUri(vaipc.PipelineBucket, vaipc.pipelineStorageObject(pipelineName, pipelineVersion))
}

type VAIProvider struct {
}

func (vaip VAIProvider) gcsClient(ctx context.Context, providerConfig VAIProviderConfig) (*storage.Client, error) {
	var client *storage.Client
	var err error
	if providerConfig.GcsEndpoint != "" {
		client, err = storage.NewClient(ctx, option.WithoutAuthentication(), option.WithEndpoint(providerConfig.GcsEndpoint))
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

	writer := client.Bucket(providerConfig.PipelineBucket).Object(providerConfig.pipelineStorageObject(id, pipelineDefinition.Version)).NewWriter(ctx)
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

func (vaip VAIProvider) createSchedulerJobPb(providerConfig VAIProviderConfig, runScheduleDefinition RunScheduleDefinition) (*schedulerpb.Job, error) {
	schedule, err := ParseCron(runScheduleDefinition.Schedule)
	if err != nil {
		return nil, err
	}

	runIntent := RunIntent{
		RunConfigurationName: runScheduleDefinition.RunConfigurationName,
		PipelineName:         runScheduleDefinition.PipelineName,
		PipelineVersion:      runScheduleDefinition.PipelineVersion,
		RuntimeParameters:    runScheduleDefinition.RuntimeParameters,
	}

	data, err := json.Marshal(runIntent)
	if err != nil {
		return nil, err
	}

	return &schedulerpb.Job{
		Name:     providerConfig.schedulerJobName(fmt.Sprintf("rc-%s", runScheduleDefinition.Name)),
		Schedule: schedule.PrintStandard(),
		Target: &schedulerpb.Job_PubsubTarget{
			PubsubTarget: &schedulerpb.PubsubTarget{
				TopicName: providerConfig.runIntentsTopicFullName(),
				Data:      data,
			},
		},
	}, nil
}

func (vaip VAIProvider) CreateRun(ctx context.Context, providerConfig VAIProviderConfig, runDefinition RunDefinition) (string, error) {
	return vaip.EnqueueRun(ctx, providerConfig, RunIntent{
		PipelineName:      runDefinition.PipelineName,
		PipelineVersion:   runDefinition.PipelineVersion,
		RuntimeParameters: runDefinition.RuntimeParameters,
		RunName:           runDefinition.Name,
	})
}

func (vaip VAIProvider) DeleteRun(_ context.Context, _ VAIProviderConfig, _ string) error {
	return nil
}

func (vaip VAIProvider) CreateRunSchedule(ctx context.Context, providerConfig VAIProviderConfig, runScheduleDefinition RunScheduleDefinition) (string, error) {
	client, err := scheduler.NewCloudSchedulerClient(ctx)
	if err != nil {
		return "", err
	}

	jobPb, err := vaip.createSchedulerJobPb(providerConfig, runScheduleDefinition)
	if err != nil {
		return "", err
	}

	job, err := client.CreateJob(ctx, &schedulerpb.CreateJobRequest{
		Parent: providerConfig.parent(),
		Job:    jobPb,
	})

	if err != nil {
		return "", err
	}

	return job.Name, nil
}

func (vaip VAIProvider) UpdateRunSchedule(ctx context.Context, providerConfig VAIProviderConfig, runScheduleDefinition RunScheduleDefinition, providerId string) (string, error) {
	client, err := scheduler.NewCloudSchedulerClient(ctx)
	if err != nil {
		return "", err
	}

	jobPb, err := vaip.createSchedulerJobPb(providerConfig, runScheduleDefinition)
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

func (vaip VAIProvider) DeleteRunSchedule(ctx context.Context, _ VAIProviderConfig, providerId string) error {
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

func (vaip VAIProvider) EnqueueRun(ctx context.Context, providerConfig VAIProviderConfig, runIntent RunIntent) (string, error) {
	pubsubClient, err := pubsub.NewClient(ctx, providerConfig.VaiProject)
	if err != nil {
		return "", err
	}

	topic := pubsubClient.Topic(providerConfig.RunsTopic)
	defer topic.Stop()

	runLabels := map[string]string{
		labels.PipelineName:    runIntent.PipelineName,
		labels.PipelineVersion: runIntent.PipelineVersion,
		labels.Namespace:       runIntent.RunName.Namespace,
	}

	var runId string

	if !runIntent.RunName.Empty() {
		runId = fmt.Sprintf(runIntent.RunName.Name)
	} else if !runIntent.RunConfigurationName.Empty() {
		runId = fmt.Sprintf("%s-%s", runIntent.RunConfigurationName.Name, uuid.New().String())
		runConfigurationName, err := runIntent.RunConfigurationName.String()
		if err != nil {
			return "", err
		}

		runLabels[labels.RunConfiguration] = runConfigurationName
	} else {
		runId = fmt.Sprintf("%s", uuid.New().String())
	}

	vaiRun := VAIRun{
		RunId:             runId,
		PipelineUri:       providerConfig.pipelineUri(runIntent.PipelineName, runIntent.PipelineVersion),
		Labels:            runLabels,
		RuntimeParameters: runIntent.RuntimeParameters,
	}

	payload, err := json.Marshal(vaiRun)
	if err != nil {
		return "", err
	}

	_, err = topic.Publish(ctx, &pubsub.Message{Data: payload}).Get(ctx)

	if err != nil {
		return "", err
	}

	return runId, nil
}

func (vaip VAIProvider) specFromTemplateUri(ctx context.Context, providerConfig VAIProviderConfig, job *aiplatformpb.PipelineJob) error {
	gcsClient, err := vaip.gcsClient(ctx, providerConfig)
	raw := map[string]interface{}{}

	r := regexp.MustCompile(`gs://([^/]+)/(.+)`)
	matched := r.FindStringSubmatch(job.TemplateUri)
	if len(matched) < 3 {
		return errors.New("invalid gs URI")
	}

	reader, err := gcsClient.Bucket(matched[1]).Object(matched[2]).NewReader(ctx)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buf.Bytes(), &raw)
	if err != nil {
		return err
	}

	pipelineSpec, err := structpb.NewStruct(raw["pipelineSpec"].(map[string]interface{}))
	if err != nil {
		return err
	}
	job.PipelineSpec = pipelineSpec

	displayName := raw["displayName"].(string)
	job.DisplayName = displayName

	labels := raw["labels"].(map[string]interface{})
	if job.Labels == nil {
		job.Labels = map[string]string{}
	}
	for k, v := range labels {
		job.Labels[k] = v.(string)
	}

	gcsOutputDirectory := raw["runtimeConfig"].(map[string]interface{})["gcsOutputDirectory"].(string)

	job.RuntimeConfig.GcsOutputDirectory = gcsOutputDirectory

	return nil
}

func (vaip VAIProvider) SubmitRun(ctx context.Context, providerConfig VAIProviderConfig, vaiRun VAIRun) error {
	pipelineClient, err := aiplatform.NewPipelineClient(ctx, option.WithEndpoint(providerConfig.vaiEndpoint()))
	if err != nil {
		return err
	}
	defer pipelineClient.Close()

	parameters := make(map[string]*aiplatformpb.Value, len(vaiRun.RuntimeParameters))
	for name, value := range vaiRun.RuntimeParameters {
		parameters[name] = &aiplatformpb.Value{
			Value: &aiplatformpb.Value_StringValue{
				StringValue: value,
			},
		}
	}

	pipelineJob := &aiplatformpb.PipelineJob{
		Labels:         vaiRun.Labels,
		TemplateUri:    vaiRun.PipelineUri,
		ServiceAccount: providerConfig.VaiJobServiceAccount,
		RuntimeConfig: &aiplatformpb.PipelineJob_RuntimeConfig{
			Parameters: parameters,
		},
	}

	err = vaip.specFromTemplateUri(ctx, providerConfig, pipelineJob)
	if err != nil {
		return err
	}

	req := &aiplatformpb.CreatePipelineJobRequest{
		Parent:        providerConfig.parent(),
		PipelineJobId: vaiRun.RunId,
		PipelineJob:   pipelineJob,
	}

	_, err = pipelineClient.CreatePipelineJob(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (vaip VAIProvider) EventingServer(ctx context.Context, providerConfig VAIProviderConfig) (generic.EventingServer, error) {
	pubSubClient, err := pubsub.NewClient(ctx, providerConfig.VaiProject)
	if err != nil {
		return nil, err
	}
	runsSubscription := pubSubClient.Subscription(providerConfig.EventsourceRunsSubscription)

	pipelineJobClient, err := aiplatform.NewPipelineClient(ctx, option.WithEndpoint(providerConfig.vaiEndpoint()))
	if err != nil {
		return nil, err
	}

	return &VaiEventingServer{
		ProviderConfig:    providerConfig,
		RunsSubscription:  runsSubscription,
		PipelineJobClient: pipelineJobClient,
		Logger:            common.LoggerFromContext(ctx),
	}, nil
}
