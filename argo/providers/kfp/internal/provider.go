package internal

import (
	"context"
	"errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_client/experiment_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_model"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/job_client/job_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/job_model"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_client/pipeline_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_client/pipeline_upload_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/run_client/run_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/run_model"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/argo/providers/base"
	"gopkg.in/yaml.v2"
	"os"
)

type KfpProviderConfig struct {
	Name       string     `yaml:"name"`
	Parameters Parameters `yaml:"parameters"`
}

type Parameters struct {
	RestKfpApiUrl            string `yaml:"restKfpApiUrl,omitempty"`
	GrpcMetadataStoreAddress string `yaml:"grpcMetadataStoreAddress,omitempty"`
	GrpcKfpApiAddress        string `yaml:"grpcKfpApiAddress,omitempty"`
}

type ResourceReferences struct {
	PipelineName         common.NamespacedName        `yaml:"pipelineName"`
	RunConfigurationName common.NamespacedName        `yaml:"runConfigurationName"`
	RunName              common.NamespacedName        `yaml:"runName"`
	Artifacts            []pipelinesv1.OutputArtifact `yaml:"artifacts,omitempty"`
}

const KfpResourceNotFoundCode = 5

type KfpProvider struct{}

func (kfpp KfpProvider) CreatePipeline(ctx context.Context, providerConfig KfpProviderConfig, pipelineDefinition PipelineDefinition, pipelineFilePath string) (string, error) {
	reader, err := os.Open(pipelineFilePath)
	if err != nil {
		return "", err
	}

	pipelineUploadService, err := pipelineUploadService(providerConfig)
	if err != nil {
		return "", err
	}

	pipelineName, err := ResourceNameFromNamespacedName(pipelineDefinition.Name)
	if err != nil {
		return "", err
	}

	result, err := pipelineUploadService.UploadPipeline(&pipeline_upload_service.UploadPipelineParams{
		Name:       &pipelineName,
		Uploadfile: runtime.NamedReader(pipelineFilePath, reader),
		Context:    ctx,
	}, nil)
	if err != nil {
		return "", err
	}

	return kfpp.UpdatePipeline(ctx, providerConfig, pipelineDefinition, result.Payload.ID, pipelineFilePath)
}

func (kfpp KfpProvider) UpdatePipeline(ctx context.Context, providerConfig KfpProviderConfig, pipelineDefinition PipelineDefinition, id string, pipelineFilePath string) (string, error) {
	reader, err := os.Open(pipelineFilePath)
	if err != nil {
		return id, err
	}

	pipelineUploadService, err := pipelineUploadService(providerConfig)
	if err != nil {
		return id, err
	}

	_, err = pipelineUploadService.UploadPipelineVersion(&pipeline_upload_service.UploadPipelineVersionParams{
		Name:       &pipelineDefinition.Version,
		Uploadfile: runtime.NamedReader(pipelineFilePath, reader),
		Pipelineid: &id,
		Context:    ctx,
	}, nil)

	return id, err
}

func (kfpp KfpProvider) DeletePipeline(ctx context.Context, providerConfig KfpProviderConfig, id string) error {
	pipelineService, err := NewPipelineService(providerConfig)
	if err != nil {
		return err
	}

	_, err = pipelineService.DeletePipeline(&pipeline_service.DeletePipelineParams{
		ID:      id,
		Context: ctx,
	}, nil)

	if err != nil {
		var deletePipelineError *pipeline_service.DeletePipelineDefault
		if errors.As(err, &deletePipelineError) {
			return ignoreNotFound(err, deletePipelineError.Payload.Code)
		}
	}

	return err
}

func (kfpp KfpProvider) CreateRun(ctx context.Context, providerConfig KfpProviderConfig, runDefinition RunDefinition) (string, error) {
	pipelineService, err := NewPipelineService(providerConfig)
	if err != nil {
		return "", err
	}

	pipelineName, err := ResourceNameFromNamespacedName(runDefinition.PipelineName)
	if err != nil {
		return "", err
	}

	pipelineId, err := pipelineService.PipelineIdForName(ctx, pipelineName)
	if err != nil {
		return "", err
	}

	jobParameters := make([]*run_model.APIParameter, 0, len(runDefinition.RuntimeParameters))
	for name, value := range runDefinition.RuntimeParameters {
		jobParameters = append(jobParameters, &run_model.APIParameter{Name: name, Value: value})
	}

	pipelineVersionId, err := pipelineService.PipelineVersionIdForName(ctx, runDefinition.PipelineVersion, pipelineId)
	if err != nil {
		return "", err
	}

	experimentService, err := NewExperimentService(providerConfig)
	if err != nil {
		return "", err
	}

	experimentVersion, err := experimentService.ExperimentIdByName(ctx, runDefinition.ExperimentName)

	runService, err := runService(providerConfig)
	if err != nil {
		return "", err
	}

	// needed to write metadata of the job as no other field is possible
	runAsDescription, err := yaml.Marshal(ResourceReferences{
		RunName:              runDefinition.Name,
		RunConfigurationName: runDefinition.RunConfigurationName,
		PipelineName:         runDefinition.PipelineName,
		Artifacts:            runDefinition.Artifacts,
	})
	if err != nil {
		return "", err
	}

	runDefinitionName, err := ResourceNameFromNamespacedName(runDefinition.Name)
	if err != nil {
		return "", err
	}

	runResult, err := runService.CreateRun(&run_service.CreateRunParams{
		Body: &run_model.APIRun{
			Name: runDefinitionName,
			PipelineSpec: &run_model.APIPipelineSpec{
				PipelineID: pipelineId,
				Parameters: jobParameters,
			},
			Description: string(runAsDescription),
			ResourceReferences: []*run_model.APIResourceReference{
				{
					Key: &run_model.APIResourceKey{
						Type: run_model.APIResourceTypeEXPERIMENT,
						ID:   experimentVersion,
					},
					Relationship: run_model.APIRelationshipOWNER,
				},
				{
					Key: &run_model.APIResourceKey{
						Type: run_model.APIResourceTypePIPELINEVERSION,
						ID:   pipelineVersionId,
					},
					Relationship: run_model.APIRelationshipCREATOR,
				},
				{
					Key: &run_model.APIResourceKey{
						Type: run_model.APIResourceTypeNAMESPACE,
						ID:   runDefinition.Name.Namespace,
					},
					Relationship: run_model.APIRelationshipOWNER,
				},
			},
		},
		Context: ctx,
	}, nil)

	if err != nil {
		return "", err
	}

	return runResult.Payload.Run.ID, nil
}

func (kfpp KfpProvider) DeleteRun(_ context.Context, _ KfpProviderConfig, _ string) error {
	return nil
}

func createAPICronSchedule(rsd RunScheduleDefinition) (*job_model.APICronSchedule, error) {
	cronExpression, err := ParseCron(rsd.Schedule.CronExpression)
	if err != nil {
		return nil, err
	}

	schedule := &job_model.APICronSchedule{
		Cron: cronExpression.PrintGo(),
	}

	if rsd.Schedule.StartTime != nil {
		schedule.StartTime = strfmt.DateTime(rsd.Schedule.StartTime.Time)
	}

	if rsd.Schedule.EndTime != nil {
		schedule.EndTime = strfmt.DateTime(rsd.Schedule.EndTime.Time)
	}

	return schedule, nil
}

func (kfpp KfpProvider) CreateRunSchedule(ctx context.Context, providerConfig KfpProviderConfig, runScheduleDefinition RunScheduleDefinition) (string, error) {
	pipelineService, err := NewPipelineService(providerConfig)
	if err != nil {
		return "", err
	}

	pipelineName, err := ResourceNameFromNamespacedName(runScheduleDefinition.PipelineName)
	if err != nil {
		return "", err
	}

	pipelineId, err := pipelineService.PipelineIdForName(ctx, pipelineName)
	if err != nil {
		return "", err
	}

	pipelineVersionId, err := pipelineService.PipelineVersionIdForName(ctx, runScheduleDefinition.PipelineVersion, pipelineId)
	if err != nil {
		return "", err
	}

	experimentService, err := NewExperimentService(providerConfig)
	if err != nil {
		return "", err
	}

	experimentVersion, err := experimentService.ExperimentIdByName(ctx, runScheduleDefinition.ExperimentName)

	apiCronSchedule, err := createAPICronSchedule(runScheduleDefinition)
	if err != nil {
		return "", err
	}

	// needed to write metadata of the job as no other field is possible
	runScheduleAsDescription, err := yaml.Marshal(ResourceReferences{
		PipelineName:         runScheduleDefinition.PipelineName,
		RunConfigurationName: runScheduleDefinition.RunConfigurationName,
		Artifacts:            runScheduleDefinition.Artifacts,
	})
	if err != nil {
		return "", err
	}

	jobService, err := jobService(providerConfig)
	if err != nil {
		return "", err
	}

	jobParameters := make([]*job_model.APIParameter, 0, len(runScheduleDefinition.RuntimeParameters))
	for name, value := range runScheduleDefinition.RuntimeParameters {
		jobParameters = append(jobParameters, &job_model.APIParameter{Name: name, Value: value})
	}

	jobName, err := ResourceNameFromNamespacedName(runScheduleDefinition.Name)
	if err != nil {
		return "", err
	}

	jobResult, err := jobService.CreateJob(&job_service.CreateJobParams{
		Body: &job_model.APIJob{
			PipelineSpec: &job_model.APIPipelineSpec{
				PipelineID: pipelineId,
				Parameters: jobParameters,
			},
			Description:    string(runScheduleAsDescription),
			Name:           jobName,
			MaxConcurrency: 1,
			Enabled:        true,
			NoCatchup:      true,
			ResourceReferences: []*job_model.APIResourceReference{
				{
					Key: &job_model.APIResourceKey{
						Type: job_model.APIResourceTypeEXPERIMENT,
						ID:   experimentVersion,
					},
					Relationship: job_model.APIRelationshipOWNER,
				},
				{
					Key: &job_model.APIResourceKey{
						Type: job_model.APIResourceTypePIPELINEVERSION,
						ID:   pipelineVersionId,
					},
					Relationship: job_model.APIRelationshipCREATOR,
				},
			},
			Trigger: &job_model.APITrigger{
				CronSchedule: apiCronSchedule,
			},
		},
		Context: ctx,
	}, nil)
	if err != nil {
		return "", err
	}

	return jobResult.Payload.ID, nil
}

func (kfpp KfpProvider) UpdateRunSchedule(ctx context.Context, providerConfig KfpProviderConfig, runScheduleDefinition RunScheduleDefinition, id string) (string, error) {
	if err := kfpp.DeleteRunSchedule(ctx, providerConfig, id); err != nil {
		return id, err
	}

	return kfpp.CreateRunSchedule(ctx, providerConfig, runScheduleDefinition)
}

func (kfpp KfpProvider) DeleteRunSchedule(ctx context.Context, providerConfig KfpProviderConfig, id string) error {
	jobService, err := jobService(providerConfig)
	if err != nil {
		return err
	}

	_, err = jobService.DeleteJob(&job_service.DeleteJobParams{
		ID:      id,
		Context: ctx,
	}, nil)

	if err != nil {
		var deleteJobError *job_service.DeleteJobDefault
		if errors.As(err, &deleteJobError) {
			return ignoreNotFound(err, deleteJobError.Payload.Code)
		}
	}

	return err
}

func (kfpp KfpProvider) CreateExperiment(ctx context.Context, providerConfig KfpProviderConfig, experimentDefinition ExperimentDefinition) (string, error) {
	experimentService, err := NewExperimentService(providerConfig)
	if err != nil {
		return "", err
	}

	experimentName, err := ResourceNameFromNamespacedName(experimentDefinition.Name)

	result, err := experimentService.CreateExperiment(&experiment_service.CreateExperimentParams{
		Body: &experiment_model.APIExperiment{
			Name:        experimentName,
			Description: experimentDefinition.Description,
		},
		Context: ctx,
	}, nil)
	if err != nil {
		return "", err
	}

	return result.Payload.ID, nil
}

func (kfpp KfpProvider) UpdateExperiment(ctx context.Context, providerConfig KfpProviderConfig, experimentDefinition ExperimentDefinition, id string) (string, error) {
	if err := kfpp.DeleteExperiment(ctx, providerConfig, id); err != nil {
		return id, err
	}

	return kfpp.CreateExperiment(ctx, providerConfig, experimentDefinition)
}

func (kfpp KfpProvider) DeleteExperiment(ctx context.Context, providerConfig KfpProviderConfig, id string) error {
	experimentService, err := NewExperimentService(providerConfig)
	if err != nil {
		return err
	}

	_, err = experimentService.DeleteExperiment(&experiment_service.DeleteExperimentParams{
		ID:      id,
		Context: ctx,
	}, nil)

	return err
}

func ignoreNotFound(err error, code int32) error {
	if code == KfpResourceNotFoundCode {
		return nil
	}
	return err
}
