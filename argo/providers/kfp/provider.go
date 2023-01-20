package kfp

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime"
	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_client/experiment_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_model"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/job_client/job_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/job_model"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_client/pipeline_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_client/pipeline_upload_service"
	"github.com/sky-uk/kfp-operator/providers/base"
	. "github.com/sky-uk/kfp-operator/providers/base"
	"github.com/sky-uk/kfp-operator/providers/base/generic"
	"github.com/sky-uk/kfp-operator/providers/kfp/ml_metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

const KfpResourceNotFoundCode = 5

type KfpProviderConfig struct {
	RestKfpApiUrl            string `yaml:"restKfpApiUrl,omitempty"`
	GrpcMetadataStoreAddress string `yaml:"grpcMetadataStoreAddress,omitempty"`
	GrpcKfpApiAddress        string `yaml:"grpcKfpApiAddress,omitempty"`
}

type KfpProvider struct{}

func (kfpp KfpProvider) CreatePipeline(ctx context.Context, providerConfig KfpProviderConfig, pipelineDefinition PipelineDefinition, pipelineFileName string) (string, error) {
	reader, err := os.Open(pipelineFileName)
	if err != nil {
		return "", err
	}

	pipelineUploadService, err := pipelineUploadService(providerConfig)
	if err != nil {
		return "", err
	}

	result, err := pipelineUploadService.UploadPipeline(&pipeline_upload_service.UploadPipelineParams{
		Name:       &pipelineDefinition.Name,
		Uploadfile: runtime.NamedReader(pipelineFileName, reader),
		Context:    ctx,
	}, nil)
	if err != nil {
		return "", err
	}

	return kfpp.UpdatePipeline(ctx, providerConfig, pipelineDefinition, result.Payload.ID, pipelineFileName)
}

func (kfpp KfpProvider) UpdatePipeline(ctx context.Context, providerConfig KfpProviderConfig, pipelineDefinition PipelineDefinition, id string, pipelineFile string) (string, error) {
	reader, err := os.Open(pipelineFile)
	if err != nil {
		return id, err
	}

	pipelineUploadService, err := pipelineUploadService(providerConfig)
	if err != nil {
		return id, err
	}

	_, err = pipelineUploadService.UploadPipelineVersion(&pipeline_upload_service.UploadPipelineVersionParams{
		Name:       &pipelineDefinition.Version,
		Uploadfile: runtime.NamedReader(pipelineFile, reader),
		Pipelineid: &id,
		Context:    ctx,
	}, nil)

	return id, err
}

func (kfpp KfpProvider) DeletePipeline(ctx context.Context, providerConfig KfpProviderConfig, id string) error {
	pipelineService, err := pipelineService(providerConfig)
	if err != nil {
		return err
	}

	_, err = pipelineService.DeletePipeline(&pipeline_service.DeletePipelineParams{
		ID:      id,
		Context: ctx,
	}, nil)

	return err
}

func (kfpp KfpProvider) CreateRunConfiguration(ctx context.Context, providerConfig KfpProviderConfig, runConfigurationDefinition RunConfigurationDefinition) (string, error) {
	pipelineService, err := pipelineService(providerConfig)
	if err != nil {
		return "", err
	}

	pipelineResult, err := pipelineService.ListPipelines(&pipeline_service.ListPipelinesParams{
		Filter:  byNameFilter(runConfigurationDefinition.PipelineName),
		Context: ctx,
	}, nil)
	if err != nil {
		return "", err
	}
	numPipelines := len(pipelineResult.Payload.Pipelines)
	if numPipelines != 1 {
		return "", fmt.Errorf("found %d pipelines, expected exactly one", numPipelines)
	}

	pipelineVersionResult, err := pipelineService.ListPipelineVersions(&pipeline_service.ListPipelineVersionsParams{
		Filter:        pipelineVersionByNameFilter(runConfigurationDefinition.PipelineVersion),
		ResourceKeyID: &pipelineResult.Payload.Pipelines[0].ID,
		Context:       ctx,
	}, nil)
	if err != nil {
		return "", err
	}
	numPipelineVersions := len(pipelineVersionResult.Payload.Versions)
	if numPipelineVersions != 1 {
		return "", fmt.Errorf("found %d pipeline versions, expected exactly one", numPipelineVersions)
	}

	experimentService, err := experimentService(providerConfig)
	if err != nil {
		return "", err
	}

	experimentResult, err := experimentService.ListExperiment(&experiment_service.ListExperimentParams{
		Filter:  byNameFilter(runConfigurationDefinition.ExperimentName),
		Context: ctx,
	}, nil)
	if err != nil {
		return "", err
	}
	numExperiments := len(experimentResult.Payload.Experiments)
	if numExperiments != 1 {
		return "", fmt.Errorf("found %d experiments, expected exactly one", numExperiments)
	}

	schedule, err := ParseCron(runConfigurationDefinition.Schedule)
	if err != nil {
		return "", err
	}

	jobService, err := jobService(providerConfig)
	if err != nil {
		return "", err
	}

	jobParameters := make([]*job_model.APIParameter, 0, len(runConfigurationDefinition.RuntimeParameters))
	for name, value := range runConfigurationDefinition.RuntimeParameters {
		jobParameters = append(jobParameters, &job_model.APIParameter{Name: name, Value: value})
	}

	jobResult, err := jobService.CreateJob(&job_service.CreateJobParams{
		Body: &job_model.APIJob{
			PipelineSpec: &job_model.APIPipelineSpec{
				PipelineID: pipelineResult.Payload.Pipelines[0].ID,
				Parameters: jobParameters,
			},
			Name:           runConfigurationDefinition.Name,
			MaxConcurrency: 1,
			Enabled:        true,
			NoCatchup:      true,
			ResourceReferences: []*job_model.APIResourceReference{
				{
					Key: &job_model.APIResourceKey{
						Type: job_model.APIResourceTypeEXPERIMENT,
						ID:   experimentResult.Payload.Experiments[0].ID,
					},
					Relationship: job_model.APIRelationshipOWNER,
				},
				{
					Key: &job_model.APIResourceKey{
						Type: job_model.APIResourceTypePIPELINEVERSION,
						ID:   pipelineVersionResult.Payload.Versions[0].ID,
					},
					Relationship: job_model.APIRelationshipCREATOR,
				},
			},
			Trigger: &job_model.APITrigger{
				CronSchedule: &job_model.APICronSchedule{
					Cron: schedule.PrintGo(),
				},
			},
		},
		Context: ctx,
	}, nil)
	if err != nil {
		return "", err
	}

	return jobResult.Payload.ID, nil
}

func (kfpp KfpProvider) UpdateRunConfiguration(ctx context.Context, providerConfig KfpProviderConfig, runConfigurationDefinition RunConfigurationDefinition, id string) (string, error) {
	if err := kfpp.DeleteRunConfiguration(ctx, providerConfig, id); err != nil {
		return id, err
	}

	return kfpp.CreateRunConfiguration(ctx, providerConfig, runConfigurationDefinition)
}

func (kfpp KfpProvider) DeleteRunConfiguration(ctx context.Context, providerConfig KfpProviderConfig, id string) error {
	jobService, err := jobService(providerConfig)
	if err != nil {
		return err
	}

	_, err = jobService.DeleteJob(&job_service.DeleteJobParams{
		ID:      id,
		Context: ctx,
	}, nil)

	if err != nil {
		errorResult := err.(*job_service.DeleteJobDefault)
		if errorResult.Payload.Code != KfpResourceNotFoundCode {
			return err
		}
	}

	return nil
}

func (kfpp KfpProvider) CreateExperiment(ctx context.Context, providerConfig KfpProviderConfig, experimentDefinition ExperimentDefinition) (string, error) {
	experimentService, err := experimentService(providerConfig)
	if err != nil {
		return "", err
	}

	result, err := experimentService.CreateExperiment(&experiment_service.CreateExperimentParams{
		Body: &experiment_model.APIExperiment{
			Name:        experimentDefinition.Name,
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
	experimentService, err := experimentService(providerConfig)
	if err != nil {
		return err
	}

	_, err = experimentService.DeleteExperiment(&experiment_service.DeleteExperimentParams{
		ID:      id,
		Context: ctx,
	}, nil)

	return err
}

func createK8sClient() (dynamic.Interface, error) {
	var k8sConfig *rest.Config
	var err error

	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if _, err := os.Stat(kubeconfigPath); err == nil {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", "")
	}

	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(k8sConfig)
}

func ConnectToMetadataStore(address string) (*GrpcMetadataStore, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GrpcMetadataStore{
		MetadataStoreServiceClient: ml_metadata.NewMetadataStoreServiceClient(conn),
	}, nil
}

func ConnectToKfpApi(address string) (*GrpcKfpApi, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GrpcKfpApi{
		RunServiceClient: go_client.NewRunServiceClient(conn),
	}, nil
}

func (kfpp KfpProvider) EventingServer(ctx context.Context, providerConfig KfpProviderConfig) (generic.EventingServer, error) {
	k8sClient, err := createK8sClient()
	if err != nil {
		return nil, err
	}

	metadataStore, err := ConnectToMetadataStore(providerConfig.GrpcMetadataStoreAddress)
	if err != nil {
		return nil, err
	}

	kfpApi, err := ConnectToKfpApi(providerConfig.GrpcKfpApiAddress)
	if err != nil {
		return nil, err
	}

	return &KfpEventingServer{
		K8sClient:     k8sClient,
		Logger:        base.LoggerFromContext(ctx),
		MetadataStore: metadataStore,
		KfpApi:        kfpApi,
	}, nil
}
