//go:build unit

package provider

import (
	"context"
	"errors"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/testutil"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
	"github.com/stretchr/testify/mock"
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

	Context("CreateJob", func() {
		It("should return a job id", func() {
			rsd := testutil.RandomRunScheduleDefinition()
			pipelineId := "pipeline-id"
			pipelineVersionId := "pipeline-version-id"
			experimentVersion := "experiment-version"

			expectedId := "expected-job-id"
			mockJobServiceClient.On(
				"CreateJob",
				// TODO: assert on a CreateJobRequest
				mock.Anything,
			).Return(&go_client.Job{Id: expectedId}, nil)

			res, err := jobService.CreateJob(
				rsd,
				pipelineId,
				pipelineVersionId,
				experimentVersion,
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(expectedId))
		})
	})

	Context("GetJob", func() {
		It("should return a job description", func() {
			expectedReq := &go_client.GetJobRequest{Id: jobId}
			desc := "description"
			mockJobServiceClient.On("GetJob", expectedReq).Return(
				&go_client.Job{Description: desc},
				nil,
			)
			res, err := jobService.GetJob(jobId)

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(desc))
		})

		When("job service client GetJob returns error", func() {
			It("should return error", func() {
				expectedReq := &go_client.GetJobRequest{Id: jobId}
				mockJobServiceClient.On("GetJob", expectedReq).Return(
					nil,
					errors.New("failed"),
				)
				res, err := jobService.GetJob(jobId)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})
	})

	Context("DeleteJob", func() {
		It("should return nil if job is deleted", func() {
			expectedReq := &go_client.DeleteJobRequest{Id: jobId}
			mockJobServiceClient.On("DeleteJob", expectedReq).Return(nil)
			err := jobService.DeleteJob(jobId)

			Expect(err).ToNot(HaveOccurred())
		})

		When("job service client DeleteJob returns gRPC NOT_FOUND", func() {
			It("should return nil", func() {
				expectedReq := &go_client.DeleteJobRequest{Id: jobId}
				mockJobServiceClient.On("DeleteJob", expectedReq).Return(
					status.Error(codes.NotFound, "not found"),
				)
				err := jobService.DeleteJob(jobId)

				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("job service client DeleteJob returns non NOT_FOUND gRPC error ", func() {
			It("should return the error", func() {
				expectedReq := &go_client.DeleteJobRequest{Id: jobId}
				mockJobServiceClient.On("DeleteJob", expectedReq).Return(
					status.Error(codes.Canceled, "not found"),
				)
				err := jobService.DeleteJob(jobId)
				Expect(err).To(HaveOccurred())
			})
		})

		When("job service client DeleteJob returns a non gRPC error ", func() {
			It("should return the error", func() {
				expectedReq := &go_client.DeleteJobRequest{Id: jobId}
				mockJobServiceClient.On("DeleteJob", expectedReq).Return(
					errors.New("failed"),
				)
				err := jobService.DeleteJob(jobId)

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
