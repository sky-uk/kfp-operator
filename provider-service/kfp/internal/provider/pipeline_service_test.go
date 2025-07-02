//go:build unit

package provider

import (
	"context"
	"errors"

	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("PipelineService", func() {
	const (
		pipelineId  = "pipeline-id"
		versionId   = "version-id"
		versionName = "version-name"
	)

	var (
		mockPipelineServiceClient mocks.MockPipelineServiceClient
		pipelineService           PipelineService
		ctx                       = context.Background()
	)

	BeforeEach(func() {
		mockPipelineServiceClient = mocks.MockPipelineServiceClient{}
		pipelineService = &DefaultPipelineService{
			&mockPipelineServiceClient,
		}
	})

	Context("DeletePipeline", func() {
		It("should not error if pipeline is deleted", func() {
			mockPipelineServiceClient.On(
				"DeletePipeline",
				&go_client.DeletePipelineRequest{
					PipelineId: pipelineId,
				},
			).Return(nil)

			err := pipelineService.DeletePipeline(ctx, pipelineId)
			Expect(err).ToNot(HaveOccurred())
		})

		When("PipelineServiceClient returns gRPC NOT_FOUND", func() {
			It("should not error", func() {
				mockPipelineServiceClient.On(
					"DeletePipeline",
					&go_client.DeletePipelineRequest{
						PipelineId: pipelineId,
					},
				).Return(status.Errorf(codes.NotFound, "resource not found"))

				err := pipelineService.DeletePipeline(ctx, pipelineId)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("PipelineServiceClient returns non NOT_FOUND gRPC error ", func() {
			It("should return the error", func() {
				mockPipelineServiceClient.On(
					"DeletePipeline",
					&go_client.DeletePipelineRequest{
						PipelineId: pipelineId,
					},
				).Return(status.Errorf(codes.Canceled, "resource not found"))

				err := pipelineService.DeletePipeline(ctx, pipelineId)
				Expect(err).To(HaveOccurred())
			})
		})

		When("PipelineServiceClient returns a non gRPC error ", func() {
			It("should return the error", func() {
				mockPipelineServiceClient.On(
					"DeletePipeline",
					&go_client.DeletePipelineRequest{
						PipelineId: pipelineId,
					},
				).Return(errors.New("failed"))

				err := pipelineService.DeletePipeline(ctx, pipelineId)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("PipelineIdForName", func() {
		It("should return the pipeline ID if exactly one pipeline is found", func() {
			expectedResult := go_client.ListPipelinesResponse{
				Pipelines: []*go_client.Pipeline{
					{PipelineId: pipelineId},
				},
			}
			mockPipelineServiceClient.On(
				"ListPipelines",
				&go_client.ListPipelinesRequest{
					Filter: util.ByNameFilter(pipelineId),
				},
			).Return(&expectedResult, nil)

			res, err := pipelineService.PipelineIdForName(ctx, pipelineId)

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(pipelineId))
		})

		When("ListPipelines errors", func() {
			It("should return the error", func() {
				mockPipelineServiceClient.On(
					"ListPipelines",
					&go_client.ListPipelinesRequest{
						Filter: util.ByNameFilter(pipelineId),
					},
				).Return(nil, errors.New("failed"))

				res, err := pipelineService.PipelineIdForName(ctx, pipelineId)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})

		When("ListPipelines returns no pipelines", func() {
			It("should return an error", func() {
				expectedResult := go_client.ListPipelinesResponse{
					Pipelines: []*go_client.Pipeline{},
				}
				mockPipelineServiceClient.On(
					"ListPipelines",
					&go_client.ListPipelinesRequest{
						Filter: util.ByNameFilter(pipelineId),
					},
				).Return(&expectedResult, nil)

				res, err := pipelineService.PipelineIdForName(ctx, pipelineId)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})

		When("ListPipelines returns more than one pipeline", func() {
			It("should return an error", func() {
				expectedResult := go_client.ListPipelinesResponse{
					Pipelines: []*go_client.Pipeline{
						{PipelineId: "one"},
						{PipelineId: "two"},
					},
				}
				mockPipelineServiceClient.On(
					"ListPipelines",
					&go_client.ListPipelinesRequest{
						Filter: util.ByNameFilter(pipelineId),
					},
				).Return(&expectedResult, nil)

				res, err := pipelineService.PipelineIdForName(ctx, pipelineId)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})
	})

	Context("PipelineVersionIdForName", func() {
		It("should return the pipeline version ID if exactly one pipeline version is found", func() {
			expectedResult := go_client.ListPipelineVersionsResponse{
				PipelineVersions: []*go_client.PipelineVersion{
					{PipelineId: versionId},
				},
			}
			mockPipelineServiceClient.On(
				"ListPipelineVersions",
				&go_client.ListPipelineVersionsRequest{
					PipelineId: pipelineId,
					Filter:     util.ByNameFilter(versionName),
				},
			).Return(&expectedResult, nil)

			res, err := pipelineService.PipelineVersionIdForName(ctx, versionName, pipelineId)

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(versionId))
		})

		When("ListPipelineVersions errors", func() {
			It("should return the error", func() {
				mockPipelineServiceClient.On(
					"ListPipelineVersions",
					&go_client.ListPipelineVersionsRequest{
						PipelineId: pipelineId,
						Filter:     util.ByNameFilter(versionName),
					},
				).Return(nil, errors.New("failed"))

				res, err := pipelineService.PipelineVersionIdForName(ctx, versionName, pipelineId)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})

		When("ListPipelineVersions returns no pipeline versions", func() {
			It("should return an error", func() {
				expectedResult := go_client.ListPipelineVersionsResponse{
					PipelineVersions: []*go_client.PipelineVersion{},
				}
				mockPipelineServiceClient.On(
					"ListPipelineVersions",
					&go_client.ListPipelineVersionsRequest{
						PipelineId: pipelineId,
						Filter:     util.ByNameFilter(versionName),
					},
				).Return(&expectedResult, nil)

				res, err := pipelineService.PipelineVersionIdForName(ctx, versionName, pipelineId)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})

		When("ListPipelineVersions returns more than one pipeline version", func() {
			It("should return an error", func() {
				expectedResult := go_client.ListPipelineVersionsResponse{
					PipelineVersions: []*go_client.PipelineVersion{
						{PipelineId: "one"},
						{PipelineId: "two"},
					},
				}
				mockPipelineServiceClient.On(
					"ListPipelineVersions",
					&go_client.ListPipelineVersionsRequest{
						PipelineId: pipelineId,
						Filter:     util.ByNameFilter(versionName),
					},
				).Return(&expectedResult, nil)

				res, err := pipelineService.PipelineVersionIdForName(ctx, versionName, pipelineId)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})
	})
})
