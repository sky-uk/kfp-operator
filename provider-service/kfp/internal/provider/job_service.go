package provider

import (
	"context"
	"fmt"
	"sort"

	"google.golang.org/grpc"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	baseResource "github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/util"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v2"
)

type JobService interface {
	CreateJob(
		ctx context.Context,
		rsd baseResource.RunScheduleDefinition,
		pipelineId string,
		pipelineVersionId string,
		experimentId string,
	) (string, error)
	GetJob(ctx context.Context, id string) (string, error)
	DeleteJob(ctx context.Context, id string) error
}

type DefaultJobService struct {
	client client.JobServiceClient
}

func NewJobService(conn *grpc.ClientConn) (JobService, error) {
	if conn == nil {
		return nil, fmt.Errorf(
			"no gRPC connection was provided to start job service",
		)
	}

	return &DefaultJobService{
		client: go_client.NewJobServiceClient(conn),
	}, nil
}

// CreateJob creates a job and returns the job result id.
func (js *DefaultJobService) CreateJob(
	ctx context.Context,
	rsd baseResource.RunScheduleDefinition,
	pipelineId string,
	pipelineVersionId string,
	experimentId string,
) (string, error) {
	// needed to write metadata of the job as no other field is possible
	runScheduleAsDescription, err := yaml.Marshal(resource.References{
		PipelineName:         rsd.PipelineName,
		RunConfigurationName: rsd.RunConfigurationName,
		Artifacts:            rsd.Artifacts,
	})
	if err != nil {
		return "", err
	}

	jobName, err := util.ResourceNameFromNamespacedName(rsd.Name)
	if err != nil {
		return "", err
	}

	jobParameters := make([]*go_client.Parameter, 0, len(rsd.Parameters))
	for name, value := range rsd.Parameters {
		jobParameters = append(jobParameters, &go_client.Parameter{Name: name, Value: value})
	}
	// Sort the parameters by name for consistent ordering
	// Solely for making testing easier.
	sort.Slice(jobParameters, func(i, j int) bool {
		return jobParameters[i].Name < jobParameters[j].Name
	})

	apiCronSchedule, err := createAPICronSchedule(rsd)
	if err != nil {
		return "", err
	}

	job, err := js.client.CreateJob(ctx, &go_client.CreateJobRequest{
		Job: &go_client.Job{
			Id:          "",
			Name:        jobName,
			Description: string(runScheduleAsDescription),
			PipelineSpec: &go_client.PipelineSpec{
				PipelineId: pipelineId,
				Parameters: jobParameters,
			},
			ResourceReferences: []*go_client.ResourceReference{
				{
					Key: &go_client.ResourceKey{
						Type: go_client.ResourceType_EXPERIMENT,
						Id:   experimentId,
					},
					Relationship: go_client.Relationship_OWNER,
				},
				{
					Key: &go_client.ResourceKey{
						Type: go_client.ResourceType_PIPELINE_VERSION,
						Id:   pipelineVersionId,
					},
					Relationship: go_client.Relationship_CREATOR,
				},
			},
			MaxConcurrency: 1,
			Trigger: &go_client.Trigger{
				Trigger: &go_client.Trigger_CronSchedule{CronSchedule: apiCronSchedule},
			},
			Enabled:   true,
			NoCatchup: true,
		},
	})
	if err != nil {
		return "", err
	}

	return job.Id, nil
}

// GetJob takes a job id and returns a job description.
func (js *DefaultJobService) GetJob(ctx context.Context, id string) (string, error) {
	job, err := js.client.GetJob(ctx, &go_client.GetJobRequest{Id: id})
	if err != nil {
		return "", err
	}

	return job.Description, nil
}

// DeleteJob deletes a job by job id. Does not error if there is no such job id.
func (js *DefaultJobService) DeleteJob(ctx context.Context, id string) error {
	_, err := js.client.DeleteJob(ctx, &go_client.DeleteJobRequest{Id: id})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			// is a gRPC error
			switch st.Code() {
			case codes.NotFound:
				return nil
			default:
				return err
			}
		}
		return err
	}

	return nil
}

func createAPICronSchedule(
	rsd baseResource.RunScheduleDefinition,
) (*go_client.CronSchedule, error) {
	cronExpression, err := util.ParseCron(rsd.Schedule.CronExpression)
	if err != nil {
		return nil, err
	}

	schedule := &go_client.CronSchedule{
		Cron: cronExpression.PrintGo(),
	}

	if rsd.Schedule.StartTime != nil {
		schedule.StartTime = timestamppb.New(rsd.Schedule.StartTime.Time)
	}

	if rsd.Schedule.EndTime != nil {
		schedule.EndTime = timestamppb.New(rsd.Schedule.EndTime.Time)
	}

	return schedule, nil
}
