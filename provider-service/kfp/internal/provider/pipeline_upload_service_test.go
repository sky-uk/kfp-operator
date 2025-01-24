//go:build unit

package provider

import (
	"bytes"
	"context"
	"errors"

	"github.com/go-openapi/runtime"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_client/pipeline_upload_service"
	"github.com/kubeflow/pipelines/backend/api/go_http_client/pipeline_upload_model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
)

var _ = Describe("DefaultPipelineUploadService", func() {
	var (
		mockPipelineUploadServiceClient mocks.MockPipelineUploadServiceClient
		pipelineUploadService           DefaultPipelineUploadService

		content      = []byte{86}
		pipelineName = "pipeline-name"
		uploadFile   = runtime.NamedReader(uploadPipelineFilePath, bytes.NewReader(content))
	)

	BeforeEach(func() {
		mockPipelineUploadServiceClient = mocks.MockPipelineUploadServiceClient{}
		pipelineUploadService = DefaultPipelineUploadService{
			context.Background(),
			&mockPipelineUploadServiceClient,
		}
	})

	Context("UploadPipeline", func() {
		When("client UploadPipeline succeeds", func() {
			It("should return the payload id", func() {
				expectedReq := &pipeline_upload_service.UploadPipelineParams{
					Name:       &pipelineName,
					Uploadfile: uploadFile,
					Context:    pipelineUploadService.ctx,
				}
				expectedId := "expected-id"
				mockPipelineUploadServiceClient.On(
					"UploadPipeline",
					expectedReq,
				).Return(
					&pipeline_upload_service.UploadPipelineOK{
						Payload: &pipeline_upload_model.APIPipeline{
							ID: expectedId,
						},
					},
					nil,
				)
				result, err := pipelineUploadService.UploadPipeline(
					content,
					pipelineName,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(expectedId))
			})
		})

		When("client UploadPipeline fails", func() {
			It("should return error", func() {
				expectedReq := &pipeline_upload_service.UploadPipelineParams{
					Name:       &pipelineName,
					Uploadfile: uploadFile,
					Context:    pipelineUploadService.ctx,
				}
				mockPipelineUploadServiceClient.On(
					"UploadPipeline",
					expectedReq,
				).Return(nil, errors.New("failed"))
				result, err := pipelineUploadService.UploadPipeline(
					content,
					pipelineName,
				)

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeEmpty())
			})
		})
	})

	Context("UploadPipelineVersion", func() {
		var (
			id      = "id"
			version = "version"
		)
		When("client UploadPipelineVersion succeeds", func() {
			It("should return the id", func() {
				expectedReq := &pipeline_upload_service.UploadPipelineVersionParams{
					Name:       &version,
					Pipelineid: &id,
					Uploadfile: uploadFile,
					Context:    pipelineUploadService.ctx,
				}
				mockPipelineUploadServiceClient.On(
					"UploadPipelineVersion",
					expectedReq,
				).Return(
					&pipeline_upload_service.UploadPipelineVersionOK{
						Payload: &pipeline_upload_model.APIPipelineVersion{},
					},
					nil,
				)
				err := pipelineUploadService.UploadPipelineVersion(
					id,
					content,
					version,
				)

				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("client UploadPipelineVersionParams fails", func() {
			It("should return error", func() {
				expectedReq := &pipeline_upload_service.UploadPipelineVersionParams{
					Name:       &version,
					Pipelineid: &id,
					Uploadfile: uploadFile,
					Context:    pipelineUploadService.ctx,
				}
				mockPipelineUploadServiceClient.On(
					"UploadPipelineVersion",
					expectedReq,
				).Return(nil, errors.New("failed"))
				err := pipelineUploadService.UploadPipelineVersion(
					id,
					content,
					version,
				)

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
