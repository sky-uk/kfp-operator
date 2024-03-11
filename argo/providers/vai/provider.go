package vai

import (
	"bytes"
	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/argoproj/argo-events/eventsources/sources/generic"
	pipelines_util "github.com/sky-uk/kfp-operator/apis/pipelines"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/argo/providers/base"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
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
}{
	PipelineName:              "pipeline-name",
	PipelineNamespace:         "pipeline-namespace",
	PipelineVersion:           "pipeline-version",
	RunConfigurationName:      "runconfiguration-name",
	RunConfigurationNamespace: "runconfiguration-namespace",
	RunName:                   "run-name",
	RunNamespace:              "run-namespace",
}

type VAIResource struct {
	Labels map[string]string `json:"labels"`
}

type VAILogEntry struct {
	Labels   map[string]string `json:"labels"`
	Resource VAIResource       `json:"resource"`
}

func extractBucketAndObjectFromGCSPath(gcsPath string) (string, string, error) {
	r := regexp.MustCompile(`gs://([^/]+)/(.+)`)
	matched := r.FindStringSubmatch(gcsPath)
	if len(matched) != 3 {
		return "", "", errors.New(fmt.Sprintf("invalid gs URI [%s]", gcsPath))
	}
	return matched[1], matched[2], nil
}

func gcsClient(ctx context.Context, providerConfig VAIProviderConfig) (*storage.Client, error) {
	var client *storage.Client
	var err error
	if providerConfig.GcsEndpoint != "" {
		client, err = storage.NewClient(ctx, option.WithoutAuthentication(), option.WithEndpoint(providerConfig.GcsEndpoint))
	} else {
		client, err = storage.NewClient(ctx)
	}

	return client, err
}

func enrichJobWithSpecFromTemplateUri(ctx context.Context, providerConfig VAIProviderConfig, job *aiplatformpb.PipelineJob) error {
	gcsClient, err := gcsClient(ctx, providerConfig)
	if err != nil {
		return err
	}

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

type VAIProvider struct {
}

func (vaip VAIProvider) CreatePipeline(ctx context.Context, providerConfig VAIProviderConfig, pipelineDefinition PipelineDefinition, pipelineFile string) (string, error) {
	if _, err := vaip.UpdatePipeline(ctx, providerConfig, pipelineDefinition, pipelineDefinition.Name, pipelineFile); err != nil {
		return "", err
	}

	return pipelineDefinition.Name, nil
}

func (vaip VAIProvider) UpdatePipeline(ctx context.Context, providerConfig VAIProviderConfig, pipelineDefinition PipelineDefinition, id string, pipelineFile string) (string, error) {
	client, err := gcsClient(ctx, providerConfig)
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
	client, err := gcsClient(ctx, providerConfig)
	if err != nil {
		return err
	}

	query := &storage.Query{Prefix: fmt.Sprintf("%s/", id)}

	it := client.Bucket(providerConfig.PipelineBucket).Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
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
	runId := runDefinition.Name.Name

	pipelineClient, err := aiplatform.NewPipelineClient(ctx, option.WithEndpoint(providerConfig.vaiEndpoint()))
	if err != nil {
		return "", err
	}
	defer pipelineClient.Close()

	parameters := make(map[string]*aiplatformpb.Value, len(runDefinition.RuntimeParameters))
	for name, value := range runDefinition.RuntimeParameters {
		parameters[name] = &aiplatformpb.Value{
			Value: &aiplatformpb.Value_StringValue{
				StringValue: value,
			},
		}
	}

	pipelineJob := &aiplatformpb.PipelineJob{
		Labels:         runLabelsFromRunDefinition(runDefinition),
		TemplateUri:    providerConfig.pipelineUri(runDefinition.PipelineName.Name, runDefinition.PipelineVersion),
		ServiceAccount: providerConfig.VaiJobServiceAccount,
		RuntimeConfig: &aiplatformpb.PipelineJob_RuntimeConfig{
			Parameters: parameters,
		},
	}

	err = enrichJobWithSpecFromTemplateUri(ctx, providerConfig, pipelineJob)
	if err != nil {
		return "", err
	}

	req := &aiplatformpb.CreatePipelineJobRequest{
		Parent:        providerConfig.parent(),
		PipelineJobId: runId,
		PipelineJob:   pipelineJob,
	}

	_, err = pipelineClient.CreatePipelineJob(ctx, req)
	if err != nil {
		return "", err
	}

	return runId, nil
}

func (vaip VAIProvider) DeleteRun(_ context.Context, _ VAIProviderConfig, _ string) error {
	return nil
}

func (vaip VAIProvider) buildPipelineJob(providerConfig VAIProviderConfig, runScheduleDefinition RunScheduleDefinition) (*aiplatformpb.PipelineJob, error) {
	parameters := make(map[string]*aiplatformpb.Value, len(runScheduleDefinition.RuntimeParameters))

	parameters = pipelines_util.MapValues(runScheduleDefinition.RuntimeParameters, func(value string) *aiplatformpb.Value {
		return &aiplatformpb.Value{
			Value: &aiplatformpb.Value_StringValue{
				StringValue: value,
			},
		}
	})

	// Note: unable to migrate from `Parameters` to `ParameterValues` at this point as `PipelineJob.pipeline_spec.schema_version` used by TFX is 2.0.0 see deprecated comment
	pipelineJob := &aiplatformpb.PipelineJob{
		Labels:         runLabelsFromSchedule(runScheduleDefinition),
		TemplateUri:    providerConfig.pipelineUri(runScheduleDefinition.PipelineName.Name, runScheduleDefinition.PipelineVersion),
		ServiceAccount: providerConfig.VaiJobServiceAccount,
		RuntimeConfig: &aiplatformpb.PipelineJob_RuntimeConfig{
			Parameters: parameters,
		},
	}

	return pipelineJob, nil
}

func (vaip VAIProvider) buildVaiScheduleFromPipelineJob(providerConfig VAIProviderConfig, runScheduleDefinition RunScheduleDefinition, pipelineJob *aiplatformpb.PipelineJob) (*aiplatformpb.Schedule, error) {
	cron, err := ParseCron(runScheduleDefinition.Schedule)
	if err != nil {
		return nil, err
	}

	return &aiplatformpb.Schedule{
		TimeSpecification: &aiplatformpb.Schedule_Cron{Cron: cron.PrintStandard()},
		Request: &aiplatformpb.Schedule_CreatePipelineJobRequest{
			CreatePipelineJobRequest: &aiplatformpb.CreatePipelineJobRequest{
				Parent:      providerConfig.parent(),
				PipelineJob: pipelineJob,
			},
		},
		DisplayName:           fmt.Sprintf("rc-%s", runScheduleDefinition.Name),
		MaxConcurrentRunCount: providerConfig.getMaxConcurrentRunCountOrDefault(),
		AllowQueueing:         true,
	}, nil
}

func (vaip VAIProvider) buildAndEnrichPipelineJob(ctx context.Context, providerConfig VAIProviderConfig, runScheduleDefinition RunScheduleDefinition) (*aiplatformpb.PipelineJob, error) {
	pipelineJob, err := vaip.buildPipelineJob(providerConfig, runScheduleDefinition)
	if err != nil {
		return nil, err
	}

	if err := enrichJobWithSpecFromTemplateUri(ctx, providerConfig, pipelineJob); err != nil {
		return nil, err
	}

	return pipelineJob, nil
}

func (vaip VAIProvider) CreateRunSchedule(ctx context.Context, providerConfig VAIProviderConfig, runScheduleDefinition RunScheduleDefinition) (string, error) {
	pipelineClient, err := aiplatform.NewScheduleClient(ctx, option.WithEndpoint(providerConfig.vaiEndpoint()))
	if err != nil {
		return "", err
	}
	defer pipelineClient.Close()

	pipelineJob, err := vaip.buildAndEnrichPipelineJob(ctx, providerConfig, runScheduleDefinition)
	if err != nil {
		return "", err
	}

	schedule, err := vaip.buildVaiScheduleFromPipelineJob(providerConfig, runScheduleDefinition, pipelineJob)
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
	pipelineClient, err := aiplatform.NewScheduleClient(ctx, option.WithEndpoint(providerConfig.vaiEndpoint()))
	if err != nil {
		return "", err
	}
	defer pipelineClient.Close()

	pipelineJob, err := vaip.buildAndEnrichPipelineJob(ctx, providerConfig, runScheduleDefinition)
	if err != nil {
		return "", err
	}

	schedule, err := vaip.buildVaiScheduleFromPipelineJob(providerConfig, runScheduleDefinition, pipelineJob)
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

func (vaip VAIProvider) DeleteRunSchedule(ctx context.Context, providerConfig VAIProviderConfig, scheduleId string) error {
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

func (vaip VAIProvider) CreateExperiment(_ context.Context, _ VAIProviderConfig, _ ExperimentDefinition) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip VAIProvider) UpdateExperiment(_ context.Context, _ VAIProviderConfig, _ ExperimentDefinition, _ string) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip VAIProvider) DeleteExperiment(_ context.Context, _ VAIProviderConfig, _ string) error {
	return errors.New("not implemented")
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

func runLabelsFromRunDefinition(runDefinition RunDefinition) map[string]string {
	runLabels := runLabelsFromPipeline(runDefinition.PipelineName, runDefinition.PipelineVersion)

	if !runDefinition.RunConfigurationName.Empty() {
		runLabels[labels.RunConfigurationName] = runDefinition.RunConfigurationName.Name
		runLabels[labels.RunConfigurationNamespace] = runDefinition.RunConfigurationName.Namespace
	}

	if !runDefinition.Name.Empty() {
		runLabels[labels.RunName] = runDefinition.Name.Name
		runLabels[labels.RunNamespace] = runDefinition.Name.Namespace
	}

	return runLabels
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
