package provider

import (
	"fmt"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	baseUtil "github.com/sky-uk/kfp-operator/provider-service/base/pkg/util"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/util"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type JobBuilder interface {
	MkRunPipelineJob(
		rd resource.RunDefinition,
	) (*aiplatformpb.PipelineJob, error)
	MkRunSchedulePipelineJob(
		rsd resource.RunScheduleDefinition,
	) (*aiplatformpb.PipelineJob, error)
	MkSchedule(
		rsd resource.RunScheduleDefinition,
		pipelineJob *aiplatformpb.PipelineJob,
		parent string,
		maxConcurrentRunCount int64,
	) (*aiplatformpb.Schedule, error)
}

type DefaultJobBuilder struct {
	serviceAccount string
	pipelineBucket string
	labelGen       LabelGen
}

// MkRunPipelineJob creates a vai pipeline job for a run that can be submitted
// to a vai pipeline job client.
func (jb DefaultJobBuilder) MkRunPipelineJob(
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

	templateUri, err := util.PipelineUri(
		rd.PipelineName,
		rd.PipelineVersion,
		jb.pipelineBucket,
	)
	if err != nil {
		return nil, err
	}

	labels, err := jb.labelGen.GenerateLabels(rd)
	if err != nil {
		return nil, err
	}

	job := &aiplatformpb.PipelineJob{
		Labels: labels,
		RuntimeConfig: &aiplatformpb.PipelineJob_RuntimeConfig{
			Parameters: params,
		},
		ServiceAccount: jb.serviceAccount,
		TemplateUri:    templateUri,
	}
	return job, nil
}

// MkRunScheudlePipelineJob creates a vai pipeline job for a run schedule that
// can be used to create a vai schedule.
func (jb DefaultJobBuilder) MkRunSchedulePipelineJob(
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

	// Note: unable to migrate from `Parameters` to `ParameterValues` at this
	// point as `PipelineJob.pipeline_spec.schema_version` used by TFX
	// is 2.0.0 see deprecated comment.
	templateUri, err := util.PipelineUri(
		rsd.PipelineName,
		rsd.PipelineVersion,
		jb.pipelineBucket,
	)
	if err != nil {
		return nil, err
	}

	labels, err := jb.labelGen.GenerateLabels(rsd)
	if err != nil {
		return nil, err
	}

	job := &aiplatformpb.PipelineJob{
		Labels:         labels,
		ServiceAccount: jb.serviceAccount,
		RuntimeConfig: &aiplatformpb.PipelineJob_RuntimeConfig{
			Parameters: params,
		},
		TemplateUri: templateUri,
	}
	return job, nil
}

// MkSchedule create a vai schedule using a vai pipeline job that can be
// submitted to a vai schedule client.
func (jb DefaultJobBuilder) MkSchedule(
	rsd resource.RunScheduleDefinition,
	pipelineJob *aiplatformpb.PipelineJob,
	parent string,
	maxConcurrentRunCount int64,
) (*aiplatformpb.Schedule, error) {
	cron, err := baseUtil.ParseCron(rsd.Schedule.CronExpression)
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
		MaxConcurrentRunCount: maxConcurrentRunCount,
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
