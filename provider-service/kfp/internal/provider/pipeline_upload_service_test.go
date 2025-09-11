//go:build unit

package provider

import (
	"bytes"
	"context"
	"errors"

	"github.com/go-openapi/runtime"
	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/pipeline_upload_client/pipeline_upload_service"
	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/pipeline_upload_model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
)

var _ = Describe("DefaultPipelineUploadService", func() {
	var (
		mockClient            mocks.MockPipelineUploadServiceClient
		pipelineUploadService DefaultPipelineUploadService

		content      = []byte{86}
		pipelineName = "pipeline-name"
		uploadFile   = runtime.NamedReader(uploadPipelineFilePath, bytes.NewReader(content))
		ctx          = context.Background()
	)

	BeforeEach(func() {
		mockClient = mocks.MockPipelineUploadServiceClient{}
		pipelineUploadService = DefaultPipelineUploadService{
			&mockClient,
		}
	})

	Context("UploadPipeline", func() {
		When("client UploadPipeline succeeds", func() {
			It("should return the payload id", func() {
				expectedReq := &pipeline_upload_service.UploadPipelineParams{
					Name:       &pipelineName,
					Uploadfile: uploadFile,
					Context:    ctx,
				}
				expectedId := "expected-id"
				mockClient.On(
					"UploadPipeline",
					expectedReq,
				).Return(
					&pipeline_upload_service.UploadPipelineOK{
						Payload: &pipeline_upload_model.V2beta1Pipeline{
							PipelineID: expectedId,
						},
					},
					nil,
				)
				result, err := pipelineUploadService.UploadPipeline(
					ctx,
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
					Context:    ctx,
				}
				mockClient.On(
					"UploadPipeline",
					expectedReq,
				).Return(nil, errors.New("failed"))
				result, err := pipelineUploadService.UploadPipeline(
					ctx,
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
					Context:    ctx,
				}
				mockClient.On(
					"UploadPipelineVersion",
					expectedReq,
				).Return(
					&pipeline_upload_service.UploadPipelineVersionOK{
						Payload: &pipeline_upload_model.V2beta1PipelineVersion{},
					},
					nil,
				)
				err := pipelineUploadService.UploadPipelineVersion(
					ctx,
					id,
					content,
					version,
				)

				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("client UploadPipelineVersion fails", func() {
			It("should return error", func() {
				expectedReq := &pipeline_upload_service.UploadPipelineVersionParams{
					Name:       &version,
					Pipelineid: &id,
					Uploadfile: uploadFile,
					Context:    ctx,
				}
				mockClient.On(
					"UploadPipelineVersion",
					expectedReq,
				).Return(nil, errors.New("failed"))
				err := pipelineUploadService.UploadPipelineVersion(
					ctx,
					id,
					content,
					version,
				)

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
