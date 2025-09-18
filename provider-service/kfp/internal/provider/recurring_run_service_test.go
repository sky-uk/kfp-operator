//go:build unit

package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/pkg/providers/base"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/testutil"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v2"
)

var _ = Describe("DefaultRecurringRunService", func() {
	var (
		mockClient          mocks.MockRecurringRunServiceClient
		mockLabelGen        mocks.MockLabelGen
		recurringRunService DefaultRecurringRunService
		rsd                 base.RunScheduleDefinition
		ctx                 = context.Background()
	)

	const (
		recurringRunId    = "recurring-run-id"
		pipelineId        = "pipeline-id"
		pipelineVersionId = "pipeline-version-id"
		experimentVersion = "experiment-version"
	)

	BeforeEach(func() {
		mockClient = mocks.MockRecurringRunServiceClient{}
		mockLabelGen = mocks.MockLabelGen{}
		recurringRunService = DefaultRecurringRunService{
			&mockClient,
			&mockLabelGen,
		}
		rsd = testutil.RandomRunScheduleDefinition()
	})

	Context("CreateRecurringRun", func() {
		It("should return a recurring run id", func() {
			rsd.Parameters = map[string]string{
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

			expectedRuntimeParams := map[string]*structpb.Value{
				"key-1": structpb.NewStringValue("value-1"),
				"key-2": structpb.NewStringValue("value-2"),
			}
			expectedCron, err := createAPICronSchedule(rsd)

			Expect(err).ToNot(HaveOccurred())

			expectedId := "expected-recurring-run-id"
			mockClient.On(
				"CreateRecurringRun",
				&go_client.CreateRecurringRunRequest{
					RecurringRun: &go_client.RecurringRun{
						RecurringRunId: "",
						DisplayName:    expectedName,
						Description:    string(expectedDescription),
						PipelineSource: &go_client.RecurringRun_PipelineVersionReference{
							PipelineVersionReference: &go_client.PipelineVersionReference{
								PipelineId:        pipelineId,
								PipelineVersionId: pipelineVersionId,
							},
						},
						RuntimeConfig: &go_client.RuntimeConfig{
							Parameters: expectedRuntimeParams,
						},
						MaxConcurrency: 1,
						Trigger: &go_client.Trigger{
							Trigger: &go_client.Trigger_CronSchedule{CronSchedule: expectedCron},
						},
						Mode:         go_client.RecurringRun_ENABLE,
						NoCatchup:    true,
						ExperimentId: experimentVersion,
					},
				},
			).Return(&go_client.RecurringRun{RecurringRunId: expectedId}, nil)
			mockLabelGen.On("GenerateLabels", mock.Anything).Return(map[string]string{}, nil)

			res, err := recurringRunService.CreateRecurringRun(
				ctx,
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
				mockLabelGen.On("GenerateLabels", mock.Anything).Return(map[string]string{}, nil)

				res, err := recurringRunService.CreateRecurringRun(
					ctx,
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
				mockLabelGen.On("GenerateLabels", mock.Anything).Return(map[string]string{}, nil)

				res, err := recurringRunService.CreateRecurringRun(
					ctx,
					rsd,
					pipelineId,
					pipelineVersionId,
					experimentVersion,
				)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})

		When("recurring run service client CreateRecurringRun returns error", func() {
			It("should return error", func() {
				mockClient.On(
					"CreateRecurringRun",
					mock.Anything,
				).Return(nil, errors.New("failed"))
				mockLabelGen.On("GenerateLabels", mock.Anything).Return(map[string]string{}, nil)

				res, err := recurringRunService.CreateRecurringRun(
					ctx,
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

	Context("GetRecurringRun", func() {
		It("should return a recurring run description", func() {
			expectedReq := &go_client.GetRecurringRunRequest{RecurringRunId: recurringRunId}
			desc := "description"
			mockClient.On("GetRecurringRun", expectedReq).Return(
				&go_client.RecurringRun{Description: desc},
				nil,
			)
			res, err := recurringRunService.GetRecurringRun(ctx, recurringRunId)

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(desc))
		})

		When("recurring run service client GetRecurringRun returns error", func() {
			It("should return error", func() {
				expectedReq := &go_client.GetRecurringRunRequest{RecurringRunId: recurringRunId}
				mockClient.On("GetRecurringRun", expectedReq).Return(
					nil,
					errors.New("failed"),
				)
				res, err := recurringRunService.GetRecurringRun(ctx, recurringRunId)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})
	})

	Context("DeleteRecurringRun", func() {
		It("should not error if recurring run is deleted", func() {
			expectedReq := &go_client.DeleteRecurringRunRequest{RecurringRunId: recurringRunId}
			mockClient.On("DeleteRecurringRun", expectedReq).Return(nil)
			err := recurringRunService.DeleteRecurringRun(ctx, recurringRunId)

			Expect(err).ToNot(HaveOccurred())
		})

		When("recurring run service client DeleteRecurringRun returns gRPC NOT_FOUND", func() {
			It("should not error", func() {
				expectedReq := &go_client.DeleteRecurringRunRequest{RecurringRunId: recurringRunId}
				mockClient.On("DeleteRecurringRun", expectedReq).Return(
					status.Error(codes.NotFound, "not found"),
				)
				err := recurringRunService.DeleteRecurringRun(ctx, recurringRunId)

				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("recurring run service client DeleteRecurringRun returns non NOT_FOUND gRPC error ", func() {
			It("should return the error", func() {
				expectedReq := &go_client.DeleteRecurringRunRequest{RecurringRunId: recurringRunId}
				mockClient.On("DeleteRecurringRun", expectedReq).Return(
					status.Error(codes.Canceled, "not found"),
				)
				err := recurringRunService.DeleteRecurringRun(ctx, recurringRunId)
				Expect(err).To(HaveOccurred())
			})
		})

		When("recurring run service client DeleteRecurringRun returns a non gRPC error ", func() {
			It("should return the error", func() {
				expectedReq := &go_client.DeleteRecurringRunRequest{RecurringRunId: recurringRunId}
				mockClient.On("DeleteRecurringRun", expectedReq).Return(
					errors.New("failed"),
				)
				err := recurringRunService.DeleteRecurringRun(ctx, recurringRunId)

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
