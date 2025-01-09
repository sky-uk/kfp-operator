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

type DefaultPipelineUploadService struct {
	restKfpApiUrl string
}

type FileHandler interface {
	Write(content []byte, pipelineName string, filePath string) (string, error)
	Update(id string, content []byte, version string, filePath string) (string, error)
	Delete(id string, pipelineName string) error
	Read(pipelineName string, filePath string) (map[string]any, error)
}

type DefaultFileHandler struct {
	ctx                   context.Context
	pipelineUploadService client.PipelineUploadService
}

func NewFileHandler(
	ctx context.Context,
	restKfpApiUrl string,
) (*DefaultFileHandler, error) {
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

	return &DefaultFileHandler{
		ctx:                   ctx,
		pipelineUploadService: pipelineUploadService,
	}, nil

}

// Write writes the ??? from ??? and returns the upload result payload ID.
func (fh *DefaultFileHandler) Write(
	content []byte,
	pipelineName string,
	pipelineFilePath string,
) (string, error) {
	reader := bytes.NewReader(content)
	uploadFile := runtime.NamedReader(pipelineFilePath, reader)
	result, err := fh.pipelineUploadService.UploadPipeline(
		&pipeline_upload_service.UploadPipelineParams{
			Name:       &pipelineName,
			Uploadfile: uploadFile,
			Context:    fh.ctx,
		},
		nil,
	)
	if err != nil {
		return "", err
	}

	return result.Payload.ID, nil
}

func (fh *DefaultFileHandler) Update(
	id string,
	content []byte,
	version string,
	pipelineFilePath string,
) (string, error) {
	reader := bytes.NewReader(content)
	uploadFile := runtime.NamedReader(pipelineFilePath, reader)
	_, err := fh.pipelineUploadService.UploadPipelineVersion(
		&pipeline_upload_service.UploadPipelineVersionParams{
			Name:       &version,
			Uploadfile: uploadFile,
			Context:    fh.ctx,
			Pipelineid: &id,
		},
		nil,
	)
	if err != nil {
		return "", err
	}

	return id, nil
}
