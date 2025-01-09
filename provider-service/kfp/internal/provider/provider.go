package provider

import (
	"context"
	"errors"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/job_client/job_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_client/pipeline_service"
	baseResource "github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/util"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v2"
)

type KfpProvider struct {
	ctx               context.Context
	config            config.KfpProviderConfig
	FileHandler       FileHandler
	pipelineService   PipelineService
	experimentService ExperimentService
	jobServiceClient  client.JobServiceClient
}

func (kfpp *KfpProvider) CreatePipeline(
	pd baseResource.PipelineDefinition,
) (string, error) {
	pipelineId, err := pd.Name.String()
	if err != nil {
		return "", err
	}

	//TODO: What should filePath be here???
	result, err := kfpp.FileHandler.Write(pd.Manifest, pipelineId, "/")
	if err != nil {
		return "", err
	}

	return kfpp.UpdatePipeline(pd, result)
}

func (kfpp *KfpProvider) UpdatePipeline(
	pd baseResource.PipelineDefinition,
	id string,
) (string, error) {
	//TODO: What should filePath be here???
	// returning a result is pointless because it's just id again. Remove?
	result, err := kfpp.FileHandler.Update(id, pd.Manifest, pd.Version, "/")
	if err != nil {
		return "", err
	}

	return result, nil
}

func (kfpp *KfpProvider) DeletePipeline(id string) error {
	_, err := kfpp.pipelineService.DeletePipeline(
		&pipeline_service.DeletePipelineParams{
			ID:      id,
			Context: kfpp.ctx,
		},
		nil,
	)

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
	pipelineName, err := util.ResourceNameFromNamespacedName(rsd.PipelineName)
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

	jobName, err := util.ResourceNameFromNamespacedName(rsd.Name)
	if err != nil {
		return "", err
	}

	jobParameters := make([]*go_client.Parameter, 0, len(rsd.RuntimeParameters))
	for name, value := range rsd.RuntimeParameters {
		jobParameters = append(jobParameters, &go_client.Parameter{Name: name, Value: value})
	}

	jobResult, err := kfpp.jobServiceClient.CreateJob(kfpp.ctx, &go_client.CreateJobRequest{
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
	id string,
) (string, error) {
	if err := kfpp.DeleteRunSchedule(id); err != nil {
		return id, err
	}

	return kfpp.CreateRunSchedule(rsd)
}

func (kfpp *KfpProvider) DeleteRunSchedule(id string) error {
	_, err := kfpp.jobServiceClient.DeleteJob(
		kfpp.ctx,
		&go_client.DeleteJobRequest{
			Id: id,
		},
		nil,
	)

	if err != nil {
		var deleteJobError *job_service.DeleteJobDefault
		if errors.As(err, &deleteJobError) {
			return ignoreNotFound(err, deleteJobError.Payload.Code)
		}
	}

	return err
}

func (kfpp *KfpProvider) CreateExperiment(
	ed baseResource.ExperimentDefinition,
) (string, error) {
	experimentName, err := util.ResourceNameFromNamespacedName(ed.Name)
	if err != nil {
		return "", err
	}

	result, err := kfpp.experimentService.CreateExperiment(
		kfpp.ctx,
		&go_client.CreateExperimentRequest{
			Experiment: &go_client.Experiment{
				Name:        experimentName,
				Description: ed.Description,
			},
		},
	)
	if err != nil {
		return "", err
	}

	return result.Id, nil
}

func (kfpp *KfpProvider) UpdateExperiment(
	ed baseResource.ExperimentDefinition,
	id string,
) (string, error) {
	if err := kfpp.DeleteExperiment(id); err != nil {
		return id, err
	}

	return kfpp.CreateExperiment(ed)
}

func (kfpp *KfpProvider) DeleteExperiment(id string) error {
	kfpp.experimentService.DeleteExperiment(
		kfpp.ctx,
		&go_client.DeleteExperimentRequest{
			Id: id,
		},
	)
	return nil
}

const KfpResourceNotFoundCode = 5

func ignoreNotFound(err error, code int32) error {
	if code == KfpResourceNotFoundCode {
		return nil
	}
	return err
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
