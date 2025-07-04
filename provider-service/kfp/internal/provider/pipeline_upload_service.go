package provider

import (
	"bytes"
	"context"
	"net/url"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/pipeline_upload_client"
	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/pipeline_upload_client/pipeline_upload_service"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
)

type PipelineUploadService interface {
	UploadPipeline(
		ctx context.Context,
		content []byte,
		pipelineName string,
	) (string, error)

	UploadPipelineVersion(
		ctx context.Context,
		id string,
		content []byte,
		version string,
	) error
}

type DefaultPipelineUploadService struct {
	// There is no gRPC client equivalent to the http client. The gRPC pipeline
	// service has similar methods (CreatePipeline, CreatePipelineVersion), but
	// require a URL to an already uploaded file rather than actually uploading
	// a file for the user.
	pipelineUploadService client.PipelineUploadService
}

const uploadPipelineFilePath string = "resource.json"

func NewPipelineUploadService(
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
		pipelineUploadService: pipelineUploadService,
	}, nil
}

// UploadPipeline uploads the compiled pipeline content and returns the upload
// result payload ID - which represents the generated pipeline id.
// pipelineFilePath file extension and content data type must align and be
// recognized by pipeline_upload_service.
func (us *DefaultPipelineUploadService) UploadPipeline(
	ctx context.Context,
	content []byte,
	pipelineName string,
) (string, error) {
	reader := bytes.NewReader(content)
	uploadFile := runtime.NamedReader(uploadPipelineFilePath, reader)
	result, err := us.pipelineUploadService.UploadPipeline(
		&pipeline_upload_service.UploadPipelineParams{
			Name:       &pipelineName,
			Uploadfile: uploadFile,
			Context:    ctx,
		},
		nil,
	)
	if err != nil {
		return "", err
	}

	return result.Payload.PipelineID, nil
}

// UploadPipelineVersion uploads the compiled pipeline content, updates the
// version of an existing pipeline id.
// pipelineFilePath file extension and content data type must align and be
// recognized by pipeline_upload_service.
func (us *DefaultPipelineUploadService) UploadPipelineVersion(
	ctx context.Context,
	id string,
	content []byte,
	version string,
) error {
	reader := bytes.NewReader(content)
	uploadFile := runtime.NamedReader(uploadPipelineFilePath, reader)
	_, err := us.pipelineUploadService.UploadPipelineVersion(
		&pipeline_upload_service.UploadPipelineVersionParams{
			Name:       &version,
			Pipelineid: &id,
			Uploadfile: uploadFile,
			Context:    ctx,
		},
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}
