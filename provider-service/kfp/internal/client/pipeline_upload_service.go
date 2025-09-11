package client

import (
	"github.com/go-openapi/runtime"
	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/pipeline_upload_client/pipeline_upload_service"
)

type PipelineUploadService interface {
	UploadPipeline(
		params *pipeline_upload_service.UploadPipelineParams,
		authInfo runtime.ClientAuthInfoWriter,
	) (*pipeline_upload_service.UploadPipelineOK, error)

	UploadPipelineVersion(
		params *pipeline_upload_service.UploadPipelineVersionParams,
		authInfo runtime.ClientAuthInfoWriter,
	) (*pipeline_upload_service.UploadPipelineVersionOK, error)
}
