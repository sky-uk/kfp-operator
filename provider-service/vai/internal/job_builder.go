package internal

import (
	"fmt"
	"strings"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type JobBuilder struct {
	serviceAccount string
	pipelineBucket string
}

func (b JobBuilder) MkRunPipelineJob(
	rd resource.RunDefinition,
) (*aiplatformpb.PipelineJob, error) {
	params := make(map[string]*aiplatformpb.Value, len(rd.RuntimeParameters))
	for name, value := range rd.RuntimeParameters {
		params[name] = &aiplatformpb.Value{
			Value: &aiplatformpb.Value_StringValue{
				StringValue: value,
			},
		}
	}

	// TODO: see if pipelinePath can be passed in instead.
	templateUri, err := b.pipelineUri(
		rd.PipelineName,
		rd.PipelineVersion,
	)
	if err != nil {
		return nil, err
	}

	job := &aiplatformpb.PipelineJob{
		Labels: b.runLabelsFromRunDefinition(rd),
		RuntimeConfig: &aiplatformpb.PipelineJob_RuntimeConfig{
			Parameters: params,
		},
		ServiceAccount: b.serviceAccount,
		TemplateUri:    templateUri,
	}
	return job, nil
}

func (b JobBuilder) MkRunSchedulePipelineJob(
	rsd resource.RunScheduleDefinition,
) (*aiplatformpb.PipelineJob, error) {
	params := make(map[string]*aiplatformpb.Value, len(rsd.RuntimeParameters))
	for name, value := range rsd.RuntimeParameters {
		params[name] = &aiplatformpb.Value{
			Value: &aiplatformpb.Value_StringValue{
				StringValue: value,
			},
		}
	}

	// Note: unable to migrate from `Parameters` to `ParameterValues` at this point as `PipelineJob.pipeline_spec.schema_version` used by TFX is 2.0.0 see deprecated comment
	templateUri, err := b.pipelineUri(rsd.PipelineName, rsd.PipelineVersion)
	if err != nil {
		return nil, err
	}

	job := &aiplatformpb.PipelineJob{
		Labels:         b.runLabelsFromSchedule(rsd),
		TemplateUri:    templateUri,
		ServiceAccount: b.serviceAccount,
		RuntimeConfig: &aiplatformpb.PipelineJob_RuntimeConfig{
			Parameters: params,
		},
	}
	return job, nil
}

func (b JobBuilder) MKSchedule(
	rsd resource.RunScheduleDefinition,
	pipelineJob *aiplatformpb.PipelineJob,
	parent string,
	maxConcurrentRunCount int64,
) (*aiplatformpb.Schedule, error) {
	cron, err := ParseCron(rsd.Schedule.CronExpression)
	if err != nil {
		return nil, err
	}

	schedule := &aiplatformpb.Schedule{
		TimeSpecification: &aiplatformpb.Schedule_Cron{Cron: cron.PrintStandard()},
		Request: &aiplatformpb.Schedule_CreatePipelineJobRequest{
			CreatePipelineJobRequest: &aiplatformpb.CreatePipelineJobRequest{
				Parent:      parent,
				PipelineJob: pipelineJob,
			},
		},
		DisplayName:           fmt.Sprintf("rc-%s-%s", rsd.Name.Namespace, rsd.Name.Name),
		MaxConcurrentRunCount: int64(maxConcurrentRunCount),
		AllowQueueing:         true,
	}

	if rsd.Schedule.StartTime != nil {
		schedule.StartTime = timestamppb.New(rsd.Schedule.StartTime.Time)
	}

	if rsd.Schedule.EndTime != nil {
		schedule.EndTime = timestamppb.New(rsd.Schedule.EndTime.Time)
	}
	return schedule, nil
}

// returns namespaceName/pipelineVersion
// e.g. namespace/name/pipelineVersion
func (b JobBuilder) pipelineStorageObject(
	pipelineName common.NamespacedName,
	pipelineVersion string,
) (string, error) {
	namespaceName, err := pipelineName.String()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", namespaceName, pipelineVersion), nil
}

func (b JobBuilder) pipelineUri(
	pipelineName common.NamespacedName,
	pipelineVersion string,
) (string, error) {
	pipelineUri, err := b.pipelineStorageObject(pipelineName, pipelineVersion)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("gs://%s/%s", b.pipelineBucket, pipelineUri), nil
}

func (b JobBuilder) runLabelsFromPipeline(
	pipelineName common.NamespacedName,
	pipelineVersion string,
) map[string]string {
	return map[string]string{
		labels.PipelineName:      pipelineName.Name,
		labels.PipelineNamespace: pipelineName.Namespace,
		labels.PipelineVersion:   strings.ReplaceAll(pipelineVersion, ".", "-"),
	}
}

func (b JobBuilder) runLabelsFromRunDefinition(
	rd resource.RunDefinition,
) map[string]string {
	runLabels := b.runLabelsFromPipeline(
		rd.PipelineName,
		rd.PipelineVersion,
	)

	if !rd.RunConfigurationName.Empty() {
		runLabels[labels.RunConfigurationName] =
			rd.RunConfigurationName.Name
		runLabels[labels.RunConfigurationNamespace] =
			rd.RunConfigurationName.Namespace
	}

	if !rd.Name.Empty() {
		runLabels[labels.RunName] = rd.Name.Name
		runLabels[labels.RunNamespace] = rd.Name.Namespace
	}

	return runLabels
}

func (b JobBuilder) runLabelsFromSchedule(
	rsd resource.RunScheduleDefinition,
) map[string]string {
	runLabels := b.runLabelsFromPipeline(rsd.PipelineName, rsd.PipelineVersion)

	if !rsd.RunConfigurationName.Empty() {
		runLabels[labels.RunConfigurationName] = rsd.RunConfigurationName.Name
		runLabels[labels.RunConfigurationNamespace] = rsd.RunConfigurationName.Namespace
	}

	return runLabels
}
