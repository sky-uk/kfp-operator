package provider

import (
	"context"
	"errors"
	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_client/pipeline_service"
	baseResource "github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	baseUtils "github.com/sky-uk/kfp-operator/provider-service/base/pkg/utils"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v2"
)

type KfpProvider struct {
	FileHandler       FileHandler
	pipelineService   PipelineService
	experimentService ExperimentService
	GrpcKfpApi        client.GrpcKfpApi
	context           context.Context
	config            config.KfpProviderConfig
}

func (kfpp *KfpProvider) CreatePipeline(
	pd baseResource.PipelineDefinition,
) (string, error) {
	manifest := []byte(pd.Manifest)

	pipelineId, err := pd.Name.String()
	if err != nil {
		return "", err
	}

	//TODO: What should filePath be here???
	if err := kfpp.FileHandler.Write(manifest, pipelineId, "/"); err != nil {
		return "", err
	}

	return "", nil
}

func (kfpp *KfpProvider) UpdatePipeline(
	pd baseResource.PipelineDefinition,
	id string,
) (string, error) {
	manifest := []byte(pd.Manifest)
	version := pd.Version

	//TODO: What should filePath be here???
	if err := kfpp.FileHandler.Update(id, manifest, version, "/"); err != nil {
		return "", err
	}

	return "", nil
}

func (kfpp *KfpProvider) DeletePipeline(id string) error {
	_, err := kfpp.pipelineService.DeletePipeline(&pipeline_service.DeletePipelineParams{
		ID:      id,
		Context: kfpp.context,
	}, nil)

	if err != nil {
		var deletePipelineError *pipeline_service.DeletePipelineDefault
		if errors.As(err, &deletePipelineError) {
			return ignoreNotFound(err, deletePipelineError.Payload.Code)
		}
	}

	return err
}

func (kfpp *KfpProvider) CreateRunSchedule(
	rsd baseResource.RunScheduleDefinition,
) (string, error) {
	pipelineName, err := baseUtils.ResourceNameFromNamespacedName(rsd.PipelineName)
	if err != nil {
		return "", err
	}

	pipelineId, err := kfpp.pipelineService.PipelineIdForName(pipelineName)
	if err != nil {
		return "", err
	}

	pipelineVersionId, err := kfpp.pipelineService.PipelineVersionIdForName(rsd.PipelineVersion, pipelineId)
	if err != nil {
		return "", err
	}

	experimentVersion, err := kfpp.experimentService.ExperimentIdByName(rsd.ExperimentName)

	apiCronSchedule, err := createAPICronSchedule(rsd)
	if err != nil {
		return "", err
	}

	// needed to write metadata of the job as no other field is possible
	runScheduleAsDescription, err := yaml.Marshal(resource.References{
		PipelineName:         rsd.PipelineName,
		RunConfigurationName: rsd.RunConfigurationName,
		Artifacts:            rsd.Artifacts,
	})
	if err != nil {
		return "", err
	}

	jobName, err := baseUtils.ResourceNameFromNamespacedName(rsd.Name)
	if err != nil {
		return "", err
	}

	jobParameters := make([]*go_client.Parameter, 0, len(rsd.RuntimeParameters))
	for name, value := range rsd.RuntimeParameters {
		jobParameters = append(jobParameters, &go_client.Parameter{Name: name, Value: value})
	}

	jobResult, err := kfpp.GrpcKfpApi.JobServiceClient.CreateJob(kfpp.context, &go_client.CreateJobRequest{
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

	if err != nil {
		return "", err
	}

	return jobResult.Id, nil
}

func (kfpp *KfpProvider) UpdateRunSchedule(
	rsd baseResource.RunScheduleDefinition,
	_ string,
) (string, error) {
	return "", nil
}

func (kfpp *KfpProvider) DeleteRunSchedule(id string) error {
	return nil
}

func (kfpp *KfpProvider) CreateExperiment(
	_ baseResource.ExperimentDefinition,
) (string, error) {
	return "", nil
}

func (kfpp *KfpProvider) UpdateExperiment(
	_ baseResource.ExperimentDefinition,
	_ string,
) (string, error) {
	return "", nil
}

func (kfpp *KfpProvider) DeleteExperiment(_ string) error {
	return nil
}

const KfpResourceNotFoundCode = 5

func ignoreNotFound(err error, code int32) error {
	if code == KfpResourceNotFoundCode {
		return nil
	}
	return err
}

func createAPICronSchedule(rsd baseResource.RunScheduleDefinition) (*go_client.CronSchedule, error) {
	cronExpression, err := utils.ParseCron(rsd.Schedule.CronExpression)
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
