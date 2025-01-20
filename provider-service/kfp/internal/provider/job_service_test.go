//go:build unit

package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	baseResource "github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/testutil"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v2"
)

var _ = Describe("DefaultJobService", func() {
	var (
		mockJobServiceClient mocks.MockJobServiceClient
		jobService           DefaultJobService
		rsd                  baseResource.RunScheduleDefinition
	)

	const (
		jobId             = "job-id"
		pipelineId        = "pipeline-id"
		pipelineVersionId = "pipeline-version-id"
		experimentVersion = "experiment-version"
	)

	BeforeEach(func() {
		mockJobServiceClient = mocks.MockJobServiceClient{}
		jobService = DefaultJobService{
			context.Background(),
			&mockJobServiceClient,
		}
		rsd = testutil.RandomRunScheduleDefinition()
	})

	Context("CreateJob", func() {
		It("should return a job id", func() {
			rsd.RuntimeParameters = map[string]string{
				"key-1": "value-1",
				"key-2": "value-2",
			}
			expectedName := fmt.Sprintf(
				"%s-%s",
				rsd.Name.Namespace,
				rsd.Name.Name,
			)
			expectedDescription, err := yaml.Marshal(resource.References{
				PipelineName:         rsd.PipelineName,
				RunConfigurationName: rsd.RunConfigurationName,
				Artifacts:            rsd.Artifacts,
			})

			Expect(err).ToNot(HaveOccurred())

			expectedRuntimeParams := []*go_client.Parameter{
				{Name: "key-1", Value: "value-1"},
				{Name: "key-2", Value: "value-2"},
			}
			expectedCron, err := createAPICronSchedule(rsd)

			Expect(err).ToNot(HaveOccurred())

			expectedId := "expected-job-id"
			mockJobServiceClient.On(
				"CreateJob",
				&go_client.CreateJobRequest{
					Job: &go_client.Job{
						Id:          "",
						Name:        expectedName,
						Description: string(expectedDescription),
						PipelineSpec: &go_client.PipelineSpec{
							PipelineId: pipelineId,
							Parameters: expectedRuntimeParams,
						},
						ResourceReferences: []*go_client.ResourceReference{
							{
								Key: &go_client.ResourceKey{
									Type: go_client.ResourceType_EXPERIMENT,
									Id:   experimentVersion,
								},
								Relationship: go_client.Relationship_OWNER,
							},
							{
								Key: &go_client.ResourceKey{
									Type: go_client.ResourceType_PIPELINE_VERSION,
									Id:   pipelineVersionId,
								},
								Relationship: go_client.Relationship_CREATOR,
							},
						},
						MaxConcurrency: 1,
						Trigger: &go_client.Trigger{
							Trigger: &go_client.Trigger_CronSchedule{CronSchedule: expectedCron},
						},
						Enabled:   true,
						NoCatchup: true,
					},
				},
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

		When("run schedule definition doesn't have a name", func() {
			It("should return error", func() {
				rsd.Name.Name = ""
				res, err := jobService.CreateJob(
					rsd,
					pipelineId,
					pipelineVersionId,
					experimentVersion,
				)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})

		When("the cron expression is invalid", func() {
			It("should return error", func() {
				rsd.Schedule.CronExpression = "invalid-cron"
				res, err := jobService.CreateJob(
					rsd,
					pipelineId,
					pipelineVersionId,
					experimentVersion,
				)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})

		When("job service CreateJob returns error", func() {
			It("should return error", func() {
				mockJobServiceClient.On(
					"CreateJob",
					mock.Anything,
				).Return(nil, errors.New("failed"))
				res, err := jobService.CreateJob(
					rsd,
					pipelineId,
					pipelineVersionId,
					experimentVersion,
				)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
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
