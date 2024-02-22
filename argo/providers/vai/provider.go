package vai

import (
	"bytes"
	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"cloud.google.com/go/pubsub"
	scheduler "cloud.google.com/go/scheduler/apiv1"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/argoproj/argo-events/eventsources/sources/generic"
	"github.com/google/uuid"
	pipelines_util "github.com/sky-uk/kfp-operator/apis/pipelines"
	pipelines "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/argo/providers/base"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	schedulerpb "google.golang.org/genproto/googleapis/cloud/scheduler/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"io"
	"os"
	"regexp"
	"strings"
)

var labels = struct {
	PipelineName              string
	PipelineNamespace         string
	PipelineVersion           string
	RunConfigurationName      string
	RunConfigurationNamespace string
	RunName                   string
	RunNamespace              string
	LegacyNamespace           string
	LegacyRunConfiguration    string
}{
	PipelineName:              "pipeline-name",
	PipelineNamespace:         "pipeline-namespace",
	PipelineVersion:           "pipeline-version",
	RunConfigurationName:      "runconfiguration-name",
	RunConfigurationNamespace: "runconfiguration-namespace",
	RunName:                   "run-name",
	RunNamespace:              "run-namespace",
	LegacyNamespace:           "namespace",
	LegacyRunConfiguration:    "run-configuration",
}

type VAIResource struct {
	Labels map[string]string `json:"labels"`
}

type VAILogEntry struct {
	Labels   map[string]string `json:"labels"`
	Resource VAIResource       `json:"resource"`
}

type VAIRun struct {
	RunId             string                     `json:"runId"`
	Labels            map[string]string          `json:"labels,omitempty"`
	PipelineUri       string                     `json:"pipelineUri"`
	RuntimeParameters map[string]string          `json:"runtimeParameters,omitempty"`
	Artifacts         []pipelines.OutputArtifact `json:"artifacts,omitempty"`
}

type RunIntent struct {
	RunConfigurationName common.NamespacedName      `json:"runConfigurationName,omitempty"`
	RunName              common.NamespacedName      `json:"runName,omitempty"`
	PipelineName         common.NamespacedName      `json:"pipelineName"`
	PipelineVersion      string                     `json:"pipelineVersion"`
	RuntimeParameters    map[string]string          `json:"runtimeParameters,omitempty"`
	Artifacts            []pipelines.OutputArtifact `json:"artifacts,omitempty"`
}

type VAIProviderConfig struct {
	Name                                  string `yaml:"name"`
	VaiProject                            string `yaml:"vaiProject"`
	VaiLocation                           string `yaml:"vaiLocation"`
	VaiJobServiceAccount                  string `yaml:"vaiJobServiceAccount"`
	GcsEndpoint                           string `yaml:"gcsEndpoint"`
	PipelineBucket                        string `yaml:"pipelineBucket"`
	RunIntentsTopic                       string `yaml:"runIntentsTopic"`
	RunsTopic                             string `yaml:"runsTopic"`
	EventsourcePipelineEventsSubscription string `yaml:"eventsourcePipelineEventsSubscription"`
	MaxConcurrentRunCount                 int    `yaml:"maxConcurrentRunCount"`
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

func (vaip VAIProvider) CreateRun(ctx context.Context, providerConfig VAIProviderConfig, runDefinition RunDefinition) (string, error) {
	return vaip.EnqueueRun(ctx, providerConfig, RunIntent{
		PipelineName:         runDefinition.PipelineName,
		PipelineVersion:      runDefinition.PipelineVersion,
		RuntimeParameters:    runDefinition.RuntimeParameters,
		RunName:              runDefinition.Name,
		RunConfigurationName: runDefinition.RunConfigurationName,
		Artifacts:            runDefinition.Artifacts,
	})
}

func (vaip VAIProvider) DeleteRun(_ context.Context, _ VAIProviderConfig, _ string) error {
	return nil
}

func buildPipelineJob(ctx context.Context, providerConfig VAIProviderConfig, runScheduleDefinition RunScheduleDefinition, mutatePipelineJob func(context.Context, VAIProviderConfig, *aiplatformpb.PipelineJob) error) (*aiplatformpb.PipelineJob, error) {
	parameters := make(map[string]*aiplatformpb.Value, len(runScheduleDefinition.RuntimeParameters))

	parameters = pipelines_util.MapValues(runScheduleDefinition.RuntimeParameters, func(value string) *aiplatformpb.Value {
		return &aiplatformpb.Value{
			Value: &aiplatformpb.Value_StringValue{
				StringValue: value,
			},
		}
	})

	// TODO migrate to non deprecated `ParameterValues` rather than `Parameters` below
	pipelineJob := &aiplatformpb.PipelineJob{
		Labels:         runLabelsFromSchedule(runScheduleDefinition),
		TemplateUri:    providerConfig.pipelineUri(runScheduleDefinition.PipelineName.Name, runScheduleDefinition.PipelineVersion),
		ServiceAccount: providerConfig.VaiJobServiceAccount,
		RuntimeConfig: &aiplatformpb.PipelineJob_RuntimeConfig{
			Parameters: parameters,
		},
	}

	if err := mutatePipelineJob(ctx, providerConfig, pipelineJob); err != nil {
		return nil, err
	}

	return pipelineJob, nil
}

func (vaip VAIProvider) buildCreatePipelineJobRequest(ctx context.Context, providerConfig VAIProviderConfig, runScheduleDefinition RunScheduleDefinition) (*aiplatformpb.Schedule_CreatePipelineJobRequest, error) {
	pipelineJob, err := buildPipelineJob(ctx, providerConfig, runScheduleDefinition, vaip.specFromTemplateUri)
	if err != nil {
		return nil, err
	}

	return &aiplatformpb.Schedule_CreatePipelineJobRequest{
		CreatePipelineJobRequest: &aiplatformpb.CreatePipelineJobRequest{
			Parent:      providerConfig.parent(),
			PipelineJob: pipelineJob,
		},
	}, nil
}

func (vaip VAIProvider) buildVaiSchedule(ctx context.Context, providerConfig VAIProviderConfig, runScheduleDefinition RunScheduleDefinition, buildPipelineJob func(context.Context, VAIProviderConfig, RunScheduleDefinition) (*aiplatformpb.Schedule_CreatePipelineJobRequest, error)) (*aiplatformpb.Schedule, error) {
	cron, err := ParseCron(runScheduleDefinition.Schedule)
	if err != nil {
		return nil, err
	}

	request, err := buildPipelineJob(ctx, providerConfig, runScheduleDefinition)
	if err != nil {
		return nil, err
	}

	return &aiplatformpb.Schedule{
		TimeSpecification:     &aiplatformpb.Schedule_Cron{Cron: cron.PrintStandard()},
		Request:               request,
		DisplayName:           fmt.Sprintf("rc-%s", runScheduleDefinition.Name),
		MaxConcurrentRunCount: int64(providerConfig.MaxConcurrentRunCount),
		AllowQueueing:         true,
	}, nil
}

func (vaip VAIProvider) CreateRunSchedule(ctx context.Context, providerConfig VAIProviderConfig, runScheduleDefinition RunScheduleDefinition) (string, error) {
	pipelineClient, err := aiplatform.NewScheduleClient(ctx, option.WithEndpoint(providerConfig.vaiEndpoint()))
	if err != nil {
		return "", err
	}
	defer pipelineClient.Close()

	schedule, err := vaip.buildVaiSchedule(ctx, providerConfig, runScheduleDefinition, vaip.buildCreatePipelineJobRequest)
	if err != nil {
		return "", err
	}

	createdSchedule, err := pipelineClient.CreateSchedule(ctx, &aiplatformpb.CreateScheduleRequest{
		Parent:   providerConfig.parent(),
		Schedule: schedule,
	})
	if err != nil {
		return "", err
	}

	return createdSchedule.Name, nil
}

func (vaip VAIProvider) UpdateRunSchedule(ctx context.Context, providerConfig VAIProviderConfig, runScheduleDefinition RunScheduleDefinition, scheduleId string) (string, error) {
	if isLegacySchedule(providerConfig, scheduleId) {
		err := vaip.DeleteLegacyRunSchedule(ctx, providerConfig, scheduleId)
		if err != nil {
			return "", err
		}
		return vaip.CreateRunSchedule(ctx, providerConfig, runScheduleDefinition)
	}

	pipelineClient, err := aiplatform.NewScheduleClient(ctx, option.WithEndpoint(providerConfig.vaiEndpoint()))
	if err != nil {
		return "", err
	}
	defer pipelineClient.Close()

	schedule, err := vaip.buildVaiSchedule(ctx, providerConfig, runScheduleDefinition, vaip.buildCreatePipelineJobRequest)
	if err != nil {
		return "", err
	}

	updateSchedule, err := pipelineClient.UpdateSchedule(ctx, &aiplatformpb.UpdateScheduleRequest{
		Schedule: schedule,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				"schedule",
			},
		},
	})
	if err != nil {
		return "", err
	}

	return updateSchedule.Name, nil
}

func (vaip VAIProvider) DeleteLegacyRunSchedule(ctx context.Context, _ VAIProviderConfig, providerId string) error {
	client, err := scheduler.NewCloudSchedulerClient(ctx)
	if err != nil {
		return err
	}

	return client.DeleteJob(ctx, &schedulerpb.DeleteJobRequest{
		Name: providerId,
	})
}

func (vaip VAIProvider) DeleteRunSchedule(ctx context.Context, providerConfig VAIProviderConfig, scheduleId string) error {
	if isLegacySchedule(providerConfig, scheduleId) {
		return vaip.DeleteLegacyRunSchedule(ctx, providerConfig, scheduleId)
	}

	pipelineClient, err := aiplatform.NewScheduleClient(ctx, option.WithEndpoint(providerConfig.vaiEndpoint()))
	if err != nil {
		return err
	}
	defer pipelineClient.Close()

	deleteSchedule, err := pipelineClient.DeleteSchedule(ctx, &aiplatformpb.DeleteScheduleRequest{
		Name: scheduleId,
	})
	if err != nil {
		return err
	}

	return deleteSchedule.Wait(ctx)
}

func isLegacySchedule(providerConfig VAIProviderConfig, scheduleId string) bool {
	return strings.HasPrefix(scheduleId, fmt.Sprintf("%s/%s", providerConfig.parent(), "jobs/"))
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

func generateRunId(runIntent RunIntent) string {
	if !runIntent.RunConfigurationName.Empty() {
		return fmt.Sprintf("%s-%s", runIntent.RunConfigurationName.Name, uuid.New().String())
	}

	if !runIntent.RunName.Empty() {
		return fmt.Sprintf(runIntent.RunName.Name)
	}

	return fmt.Sprintf("%s", uuid.New().String())
}

func runLabelsFromPipeline(pipelineName common.NamespacedName, pipelineVersion string) map[string]string {
	return map[string]string{
		labels.PipelineName:      pipelineName.Name,
		labels.PipelineNamespace: pipelineName.Namespace,
		labels.PipelineVersion:   strings.ReplaceAll(pipelineVersion, ".", "-"),
	}
}

func runLabelsFromSchedule(runScheduleDefinition RunScheduleDefinition) map[string]string {
	runLabels := runLabelsFromPipeline(runScheduleDefinition.PipelineName, runScheduleDefinition.PipelineVersion)

	if !runScheduleDefinition.RunConfigurationName.Empty() {
		runLabels[labels.RunConfigurationName] = runScheduleDefinition.RunConfigurationName.Name
		runLabels[labels.RunConfigurationNamespace] = runScheduleDefinition.RunConfigurationName.Namespace
	}

	return runLabels
}

func runLabelsFromRun(runIntent RunIntent) map[string]string {
	runLabels := runLabelsFromPipeline(runIntent.PipelineName, runIntent.PipelineVersion)

	if !runIntent.RunConfigurationName.Empty() {
		runLabels[labels.RunConfigurationName] = runIntent.RunConfigurationName.Name
		runLabels[labels.RunConfigurationNamespace] = runIntent.RunConfigurationName.Namespace
	}

	if !runIntent.RunName.Empty() {
		runLabels[labels.RunName] = runIntent.RunName.Name
		runLabels[labels.RunNamespace] = runIntent.RunName.Namespace
	}

	return runLabels
}

func (vaip VAIProvider) EnqueueRun(ctx context.Context, providerConfig VAIProviderConfig, runIntent RunIntent) (string, error) {
	pubsubClient, err := pubsub.NewClient(ctx, providerConfig.VaiProject)
	if err != nil {
		return "", err
	}

	topic := pubsubClient.Topic(providerConfig.RunsTopic)
	defer topic.Stop()

	vaiRun := VAIRun{
		RunId:             generateRunId(runIntent),
		PipelineUri:       providerConfig.pipelineUri(runIntent.PipelineName.Name, runIntent.PipelineVersion),
		Labels:            runLabelsFromRun(runIntent),
		RuntimeParameters: runIntent.RuntimeParameters,
		Artifacts:         runIntent.Artifacts,
	}

	payload, err := json.Marshal(vaiRun)
	if err != nil {
		return "", err
	}

	_, err = topic.Publish(ctx, &pubsub.Message{Data: payload}).Get(ctx)

	if err != nil {
		return "", err
	}

	return vaiRun.RunId, nil
}

func extractBucketAndObjectFromGCSPath(gcsPath string) (string, string, error) {
	r := regexp.MustCompile(`gs://([^/]+)/(.+)`)
	matched := r.FindStringSubmatch(gcsPath)
	if len(matched) < 3 {
		return "", "", errors.New(fmt.Sprintf("invalid gs URI %s", gcsPath))
	}
	return matched[1], matched[2], nil
}

func (vaip VAIProvider) specFromTemplateUri(ctx context.Context, providerConfig VAIProviderConfig, job *aiplatformpb.PipelineJob) error {
	gcsClient, err := vaip.gcsClient(ctx, providerConfig)
	raw := map[string]interface{}{}

	gcsBucket, gcsPath, err := extractBucketAndObjectFromGCSPath(job.TemplateUri)
	if err != nil {
		return err
	}

	reader, err := gcsClient.Bucket(gcsBucket).Object(gcsPath).NewReader(ctx)
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
	k8sClient, err := CreateK8sClient()
	if err != nil {
		return nil, err
	}

	pubSubClient, err := pubsub.NewClient(ctx, providerConfig.VaiProject)
	if err != nil {
		return nil, err
	}
	runsSubscription := pubSubClient.Subscription(providerConfig.EventsourcePipelineEventsSubscription)

	pipelineJobClient, err := aiplatform.NewPipelineClient(ctx, option.WithEndpoint(providerConfig.vaiEndpoint()))
	if err != nil {
		return nil, err
	}

	return &VaiEventingServer{
		K8sApi:            K8sApi{K8sClient: k8sClient},
		ProviderConfig:    providerConfig,
		RunsSubscription:  runsSubscription,
		PipelineJobClient: pipelineJobClient,
		Logger:            common.LoggerFromContext(ctx),
	}, nil
}
