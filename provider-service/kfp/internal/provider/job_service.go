package provider

import (
	"context"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	baseResource "github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/util"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v2"
)

type JobService interface {
	CreateJob() (string, error)
	GetJob(id string) (string, error)
	DeleteJob(id string) error
}

type DefaultJobService struct {
	ctx    context.Context
	client client.JobServiceClient
}

// CreateJob creates a job and returns the job result id.
func (js *DefaultJobService) CreateJob(
	rsd baseResource.RunScheduleDefinition,
	pipelineId string,
	pipelineVersionId string,
	experimentVersion string,
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

	jobParameters := make([]*go_client.Parameter, 0, len(rsd.RuntimeParameters))
	for name, value := range rsd.RuntimeParameters {
		jobParameters = append(jobParameters, &go_client.Parameter{Name: name, Value: value})
	}

	apiCronSchedule, err := createAPICronSchedule(rsd)
	if err != nil {
		return "", err
	}

	job, err := js.client.CreateJob(js.ctx, &go_client.CreateJobRequest{
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
						Id:   experimentVersion,
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

	return job.Id, nil
}

// GetJob takes a job id and returns a job description.
func (js *DefaultJobService) GetJob(id string) (string, error) {
	job, err := js.client.GetJob(js.ctx, &go_client.GetJobRequest{Id: id})
	if err != nil {
		return "", err
	}

	return job.Description, nil
}

func (js *DefaultJobService) DeleteJob(id string) error {
	_, err := js.client.DeleteJob(js.ctx, &go_client.DeleteJobRequest{Id: id})
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
