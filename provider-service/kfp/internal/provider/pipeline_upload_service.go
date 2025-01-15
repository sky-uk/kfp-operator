package provider

import (
	"bytes"
	"context"
	"net/url"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_client"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_client/pipeline_upload_service"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
)

type PipelineUploadService interface {
	UploadPipeline(
		content []byte,
		pipelineName string,
		filePath string,
	) (string, error)

	UploadPipelineVersion(
		id string,
		content []byte,
		version string,
		filePath string,
	) error
}

type DefaultPipelineUploadService struct {
	ctx                   context.Context
	pipelineUploadService client.PipelineUploadService
}

func NewPipelineUploadService(
	ctx context.Context,
	restKfpApiUrl string,
) (*DefaultPipelineUploadService, error) {
	apiUrl, err := url.Parse(restKfpApiUrl)
	if err != nil {
		return nil, err
	}

	pipelineUploadService := pipeline_upload_client.NewHTTPClientWithConfig(
		strfmt.NewFormats(),
		&pipeline_upload_client.TransportConfig{
			Host:     apiUrl.Host,
			Schemes:  []string{apiUrl.Scheme},
			BasePath: apiUrl.Path,
		},
	).PipelineUploadService

	return &DefaultPipelineUploadService{
		ctx:                   ctx,
		pipelineUploadService: pipelineUploadService,
	}, nil
}

// UploadPipeline writes the ??? from ??? and returns the upload result payload ID.
func (us *DefaultPipelineUploadService) UploadPipeline(
	content []byte,
	pipelineName string,
	pipelineFilePath string,
) (string, error) {
	reader := bytes.NewReader(content)
	uploadFile := runtime.NamedReader(pipelineFilePath, reader)
	result, err := us.pipelineUploadService.UploadPipeline(
		&pipeline_upload_service.UploadPipelineParams{
			Name:       &pipelineName,
			Uploadfile: uploadFile,
			Context:    us.ctx,
		},
		nil,
	)
	if err != nil {
		return "", err
	}

	return result.Payload.ID, nil
}

func (us *DefaultPipelineUploadService) UploadPipelineVersion(
	id string,
	content []byte,
	version string,
	pipelineFilePath string,
) error {
	reader := bytes.NewReader(content)
	uploadFile := runtime.NamedReader(pipelineFilePath, reader)
	_, err := us.pipelineUploadService.UploadPipelineVersion(
		&pipeline_upload_service.UploadPipelineVersionParams{
			Name:       &version,
			Pipelineid: &id,
			Uploadfile: uploadFile,
			Context:    us.ctx,
		},
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}
