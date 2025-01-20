package provider

import (
	"context"
	"errors"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("DefaultJobService", func() {
	var (
		mockJobServiceClient mocks.MockJobServiceClient
		jobService           DefaultJobService
		jobId                = "job-id"
	)

	BeforeEach(func() {
		mockJobServiceClient = mocks.MockJobServiceClient{}
		jobService = DefaultJobService{
			context.Background(),
			&mockJobServiceClient,
		}
	})

	Context("DeleteJob", func() {
		It("should return nil if job is deleted", func() {
			expectedReq := &go_client.DeleteJobRequest{
				Id: jobId,
			}
			mockJobServiceClient.On("DeleteJob", expectedReq).Return(nil)
			err := jobService.DeleteJob(jobId)

			Expect(err).ToNot(HaveOccurred())
		})

		When("DeleteJob returns gRPC NOT_FOUND", func() {
			It("should return nil", func() {
				expectedReq := &go_client.DeleteJobRequest{
					Id: jobId,
				}
				mockJobServiceClient.On("DeleteJob", expectedReq).Return(
					status.Error(codes.NotFound, "not found"),
				)
				err := jobService.DeleteJob(jobId)

				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("DeleteJob returns non NOT_FOUND gRPC error ", func() {
			It("should return the error", func() {
				expectedReq := &go_client.DeleteJobRequest{
					Id: jobId,
				}
				mockJobServiceClient.On("DeleteJob", expectedReq).Return(
					status.Error(codes.Canceled, "not found"),
				)
				err := jobService.DeleteJob(jobId)
				Expect(err).To(HaveOccurred())
			})
		})
		When("DeleteJob returns a non gRPC error ", func() {
			It("should return the error", func() {
				expectedReq := &go_client.DeleteJobRequest{
					Id: jobId,
				}
				mockJobServiceClient.On("DeleteJob", expectedReq).Return(
					errors.New("failed"),
				)
				err := jobService.DeleteJob(jobId)

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
