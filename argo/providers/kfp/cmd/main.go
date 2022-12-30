package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_client/experiment_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/experiment_model"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/job_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/job_client/job_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/job_model"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_client/pipeline_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_client/pipeline_upload_service"
	"github.com/sky-uk/kfp-operator/providers/base"
	. "github.com/sky-uk/kfp-operator/providers/base"
	"github.com/sky-uk/kfp-operator/providers/base/generic"
	. "github.com/sky-uk/kfp-operator/providers/kfp"
	"github.com/sky-uk/kfp-operator/providers/kfp/ml_metadata"
	"github.com/yalp/jsonpath"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"regexp"
)

const KfpResourceNotFoundCode = 5

type KfpProviderConfig struct {
	RestKfpApiUrl            string `yaml:"restKfpApiUrl,omitempty"`
	GrpcMetadataStoreAddress string `yaml:"grpcMetadataStoreAddress,omitempty"`
	GrpcKfpApiAddress        string `yaml:"grpcKfpApiAddress,omitempty"`
}

type KfpProvider struct {
	PipelineUpload pipeline_upload_client.PipelineUpload
	Pipeline       pipeline_client.Pipeline
	Job            job_client.Job
	Experiment     experiment_client.Experiment
}

func main() {
	app := NewProviderApp[KfpProviderConfig]()
	app.Run(KfpProvider{})
}

func pipelineUploadService(providerConfig KfpProviderConfig) *pipeline_upload_service.Client {
	return pipeline_upload_client.NewHTTPClientWithConfig(strfmt.NewFormats(), &pipeline_upload_client.TransportConfig{
		Host:    providerConfig.RestKfpApiUrl,
		Schemes: []string{"http"},
	}).PipelineUploadService
}

func pipelineService(providerConfig KfpProviderConfig) *pipeline_service.Client {
	return pipeline_client.NewHTTPClientWithConfig(strfmt.NewFormats(), &pipeline_client.TransportConfig{
		Host:    providerConfig.RestKfpApiUrl,
		Schemes: []string{"http"},
	}).PipelineService
}

func experimentService(providerConfig KfpProviderConfig) *experiment_service.Client {
	return experiment_client.NewHTTPClientWithConfig(strfmt.NewFormats(), &experiment_client.TransportConfig{
		Host:    providerConfig.RestKfpApiUrl,
		Schemes: []string{"http"},
	}).ExperimentService
}

func jobService(providerConfig KfpProviderConfig) *job_service.Client {
	return job_client.NewHTTPClientWithConfig(strfmt.NewFormats(), &job_client.TransportConfig{
		Host:    providerConfig.RestKfpApiUrl,
		Schemes: []string{"http"},
	}).JobService
}

func (kfpp KfpProvider) CreatePipeline(ctx context.Context, providerConfig KfpProviderConfig, pipelineDefinition PipelineDefinition, pipelineFileName string) (string, error) {
	reader, err := os.Open(pipelineFileName)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	result, err := pipelineUploadService(providerConfig).UploadPipeline(&pipeline_upload_service.UploadPipelineParams{
		Name:       &pipelineDefinition.Name,
		Uploadfile: runtime.NamedReader(pipelineFileName, reader),
		Context:    ctx,
	}, nil)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return kfpp.UpdatePipeline(ctx, providerConfig, pipelineDefinition, result.Payload.ID, pipelineFileName)
}

func (kfpp KfpProvider) UpdatePipeline(ctx context.Context, providerConfig KfpProviderConfig, pipelineDefinition PipelineDefinition, id string, pipelineFile string) (string, error) {
	reader, err := os.Open(pipelineFile)
	if err != nil {
		return "", err
	}

	_, err = pipelineUploadService(providerConfig).UploadPipelineVersion(&pipeline_upload_service.UploadPipelineVersionParams{
		Name:       &pipelineDefinition.Version,
		Uploadfile: runtime.NamedReader(pipelineFile, reader),
		Pipelineid: &id,
		Context:    ctx,
	}, nil)
	if err != nil {
		return "", err
	}

	return id, err
}

func (kfpp KfpProvider) DeletePipeline(ctx context.Context, providerConfig KfpProviderConfig, id string) error {
	_, err := pipelineService(providerConfig).DeletePipeline(&pipeline_service.DeletePipelineParams{
		ID:      id,
		Context: ctx,
	}, nil)

	return err
}

func (kfpp KfpProvider) CreateRunConfiguration(ctx context.Context, providerConfig KfpProviderConfig, runConfigurationDefinition RunConfigurationDefinition) (string, error) {
	pipelineService := pipelineService(providerConfig)
	pipelineResult, err := pipelineService.GetPipelineByName(&pipeline_service.GetPipelineByNameParams{
		Name:    runConfigurationDefinition.PipelineName,
		Context: ctx,
	}, nil)
	if err != nil {
		fmt.Println(1)
		fmt.Println(err)
		return "", err
	}

	pipelineVersionByNameFilter := fmt.Sprintf(`
	"predicates": [{
	"op": 1,
	"key": "name",
	"stringValue": %s,
	}]
	`, runConfigurationDefinition.PipelineVersion)
	pipelineVersionResult, err := pipelineService.ListPipelineVersions(&pipeline_service.ListPipelineVersionsParams{
		Filter:  &pipelineVersionByNameFilter,
		Context: ctx,
	}, nil)
	if err != nil {
		fmt.Println(2)
		fmt.Println(err)
		return "", err
	}
	if len(pipelineVersionResult.Payload.Versions) != 1 {
		fmt.Println(3)
		return "", errors.New("")
	}

	experimentByNameFilter := fmt.Sprintf(`
	"predicates": [{
	"op": 1,
	"key": "name",
	"stringValue": %s,
	}]
	`, runConfigurationDefinition.ExperimentName)
	experimentResult, err := experimentService(providerConfig).ListExperiment(&experiment_service.ListExperimentParams{
		Filter:  &experimentByNameFilter,
		Context: ctx,
	}, nil)
	if err != nil {
		fmt.Println(4)
		fmt.Println(err)
		return "", err
	}
	if len(experimentResult.Payload.Experiments) != 1 {
		fmt.Println(5)
		fmt.Println("XXX")
		return "", errors.New("")
	}

	schedule, err := ParseCron(runConfigurationDefinition.Schedule)
	if err != nil {
		fmt.Println(6)
		fmt.Println(err)
		return "", err
	}

	jobResult, err := jobService(providerConfig).CreateJob(&job_service.CreateJobParams{
		Body: &job_model.APIJob{
			PipelineSpec: &job_model.APIPipelineSpec{
				PipelineID: pipelineResult.Payload.ID,
			},
			Name: runConfigurationDefinition.Name,
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
						ID:   pipelineVersionResult.Payload.Versions[1].ID,
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
	}, nil)
	if err != nil {
		fmt.Println(7)
		fmt.Println(err)
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
	result, err := jobService(providerConfig).DeleteJob(&job_service.DeleteJobParams{
		ID:      id,
		Context: ctx,
	}, nil)
	if err != nil {
		re := regexp.MustCompile(`(?m)^.*HTTP response body: ({.*})$`)
		matches := re.FindStringSubmatch(result.Error())

		if len(matches) < 2 {
			return err
		}

		var jsonResponse interface{}
		if err = json.Unmarshal([]byte(matches[1]), &jsonResponse); err != nil {
			return err
		}

		errorCode, err := jsonpath.Read(jsonResponse, `$["code"]`)
		if err != nil {
			return err
		}

		if int(errorCode.(float64)) != KfpResourceNotFoundCode {
			return err
		}
	}

	return nil
}

func (kfpp KfpProvider) CreateExperiment(ctx context.Context, providerConfig KfpProviderConfig, experimentDefinition ExperimentDefinition) (string, error) {
	result, err := experimentService(providerConfig).CreateExperiment(&experiment_service.CreateExperimentParams{
		Body: &experiment_model.APIExperiment{
			Name:        experimentDefinition.Name,
			Description: experimentDefinition.Description,
		},
		Context: ctx,
	}, nil)
	if err != nil {
		fmt.Println(err)
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
	_, err := experimentService(providerConfig).DeleteExperiment(&experiment_service.DeleteExperimentParams{
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
