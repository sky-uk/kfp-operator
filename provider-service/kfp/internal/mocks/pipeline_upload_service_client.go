package mocks

import (
	"github.com/go-openapi/runtime"
	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/pipeline_upload_client/pipeline_upload_service"
	"github.com/stretchr/testify/mock"
)

type MockPipelineUploadServiceClient struct {
	mock.Mock
}

func (m *MockPipelineUploadServiceClient) UploadPipeline(
	in *pipeline_upload_service.UploadPipelineParams,
	_ runtime.ClientAuthInfoWriter,
) (*pipeline_upload_service.UploadPipelineOK, error) {
	args := m.Called(in)
	var response *pipeline_upload_service.UploadPipelineOK
	if arg0 := args.Get(0); arg0 != nil {
		response = arg0.(*pipeline_upload_service.UploadPipelineOK)
	}
	return response, args.Error(1)
}

func (m *MockPipelineUploadServiceClient) UploadPipelineVersion(
	in *pipeline_upload_service.UploadPipelineVersionParams,
	_ runtime.ClientAuthInfoWriter,
) (*pipeline_upload_service.UploadPipelineVersionOK, error) {
	args := m.Called(in)
	var response *pipeline_upload_service.UploadPipelineVersionOK
	if arg0 := args.Get(0); arg0 != nil {
		response = arg0.(*pipeline_upload_service.UploadPipelineVersionOK)
	}
	return response, args.Error(1)
}
