//go:build unit

package provider

import (
	"bytes"
	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/testutil"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/config"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/mocks"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var _ = Describe("Provider", func() {
	var (
		mockFileHandler    mocks.MockFileHandler
		mockPipelineClient mocks.MockPipelineJobClient
		mockScheduleClient mocks.MockScheduleClient
		mockJobBuilder     mocks.MockJobBuilder
		mockJobEnricher    mocks.MockJobEnricher
		vaiProvider        VAIProvider
		ctx                = context.Background()
	)

	BeforeEach(func() {
		mockFileHandler = mocks.MockFileHandler{}
		mockPipelineClient = mocks.MockPipelineJobClient{}
		mockScheduleClient = mocks.MockScheduleClient{}
		mockJobBuilder = mocks.MockJobBuilder{}
		mockJobEnricher = mocks.MockJobEnricher{}
		vaiProvider = VAIProvider{
			config:         &config.VAIProviderConfig{},
			fileHandler:    &mockFileHandler,
			pipelineClient: &mockPipelineClient,
			scheduleClient: &mockScheduleClient,
			jobBuilder:     &mockJobBuilder,
			jobEnricher:    &mockJobEnricher,
		}
	})

	Context("CreatePipeline", func() {
		When("creating a pipeline", func() {
			It("should return the pipeline ID", func() {
				mockFileHandler.On(
					"Write",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				pdw := testutil.RandomPipelineDefinitionWrapper()
				pid, err := vaiProvider.CreatePipeline(ctx, pdw)

				Expect(err).ToNot(HaveOccurred())
				Expect(pid).To(Equal(fmt.Sprintf(
					"%s/%s",
					pdw.PipelineDefinition.Name.Namespace,
					pdw.PipelineDefinition.Name.Name,
				)))
			})

			It("return an error when the file handler write fails", func() {
				mockFileHandler.On(
					"Write",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("failed"))

				pdw := testutil.RandomPipelineDefinitionWrapper()
				_, err := vaiProvider.CreatePipeline(ctx, pdw)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})

	Context("UpdatePipeline", func() {
		When("updating a pipeline", func() {
			It("should return the pipeline ID", func() {
				pdw := testutil.RandomPipelineDefinitionWrapper()
				mockFileHandler.On(
					"Write",
					mock.MatchedBy(func(j json.RawMessage) bool {
						return bytes.Equal(j, pdw.CompiledPipeline)
					}),
					vaiProvider.config.Parameters.PipelineBucket,
					fmt.Sprintf(
						"%s/%s/%s",
						pdw.PipelineDefinition.Name.Namespace,
						pdw.PipelineDefinition.Name.Name,
						pdw.PipelineDefinition.Version,
					),
				).Return(nil)

				pid, err := vaiProvider.UpdatePipeline(ctx, pdw, "")

				Expect(err).ToNot(HaveOccurred())
				Expect(pid).To(Equal(fmt.Sprintf(
					"%s/%s", pdw.PipelineDefinition.Name.Namespace, pdw.PipelineDefinition.Name.Name,
				)))
			})

			It("return an error when the file handler write fails", func() {
				pdw := testutil.RandomPipelineDefinitionWrapper()
				mockFileHandler.On(
					"Write",
					mock.Anything,
					vaiProvider.config.Parameters.PipelineBucket,
					mock.Anything,
				).Return(errors.New("failed"))

				_, err := vaiProvider.CreatePipeline(ctx, pdw)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})

	Context("DeletePipeline", func() {
		When("deleting a pipeline", func() {
			It("should not return an error", func() {
				id := "id"
				mockFileHandler.On(
					"Delete",
					id,
					vaiProvider.config.Parameters.PipelineBucket,
				).Return(nil)
				err := vaiProvider.DeletePipeline(ctx, id)
				Expect(err).ToNot(HaveOccurred())
			})

			It("return an error when the file handler delete fails", func() {
				mockFileHandler.On(
					"Delete",
					"pipelineId",
					vaiProvider.config.Parameters.PipelineBucket,
				).Return(errors.New("failed"))
				err := vaiProvider.DeletePipeline(ctx, "pipelineId")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})

	Context("CreateRun", func() {
		When("creating a run", func() {
			It("return a run ID", func() {
				rd := testutil.RandomRunDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					fmt.Sprintf(
						"%s/%s/%s",
						rd.PipelineName.Namespace,
						rd.PipelineName.Name,
						rd.PipelineVersion,
					),
				).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunPipelineJob", rd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(&pj, nil)
				mockPipelineClient.On(
					"CreatePipelineJob",
					&aiplatformpb.CreatePipelineJobRequest{
						Parent:        vaiProvider.config.Parent(),
						PipelineJobId: fmt.Sprintf("%s-%s-%s", rd.Name.Namespace, rd.Name.Name, rd.Version),
						PipelineJob:   &pj,
					},
				).Return(&pj, nil)
				runId, err := vaiProvider.CreateRun(ctx, rd)

				Expect(err).ToNot(HaveOccurred())
				Expect(runId).To(Equal(fmt.Sprintf("%s-%s", rd.Name.Namespace, rd.Name.Name)))
			})

			It("return an error when the file handler read fails", func() {
				rd := testutil.RandomRunDefinition()
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					mock.Anything,
				).Return(map[string]any{}, errors.New("failed"))
				_, err := vaiProvider.CreateRun(ctx, rd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job builder fails", func() {
				rd := testutil.RandomRunDefinition()
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					mock.Anything,
				).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunPipelineJob", rd).Return(nil, errors.New("failed"))
				_, err := vaiProvider.CreateRun(ctx, rd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job enricher fails", func() {
				rd := testutil.RandomRunDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					mock.Anything,
				).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunPipelineJob", rd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(nil, errors.New("failed"))
				_, err := vaiProvider.CreateRun(ctx, rd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the pipeline client fails", func() {
				rd := testutil.RandomRunDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					mock.Anything,
				).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunPipelineJob", rd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(&pj, nil)
				mockPipelineClient.On("CreatePipelineJob", mock.Anything).Return(nil, errors.New("failed"))
				_, err := vaiProvider.CreateRun(ctx, rd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})

	Context("CreateRunSchedule", func() {
		When("creating a run schedule", func() {
			It("returns a schedule name", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				pj := aiplatformpb.PipelineJob{}
				schedule := aiplatformpb.Schedule{}
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					fmt.Sprintf(
						"%s/%s/%s",
						rsd.PipelineName.Namespace,
						rsd.PipelineName.Name,
						rsd.PipelineVersion,
					),
				).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(&pj, nil)
				mockJobBuilder.On(
					"MkSchedule",
					rsd,
					&pj,
					vaiProvider.config.Parent(),
					vaiProvider.config.GetMaxConcurrentRunCountOrDefault(),
				).Return(&schedule, nil)
				mockScheduleClient.On(
					"CreateSchedule",
					&aiplatformpb.CreateScheduleRequest{
						Parent:   vaiProvider.config.Parent(),
						Schedule: &schedule,
					},
				).Return(&schedule, nil)
				scheduleName, err := vaiProvider.CreateRunSchedule(ctx, rsd)

				Expect(err).ToNot(HaveOccurred())
				Expect(scheduleName).To(Equal(schedule.Name))
			})

			It("return an error when the file handler read fails", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					mock.Anything,
				).Return(map[string]any{}, errors.New("failed"))
				_, err := vaiProvider.CreateRunSchedule(ctx, rsd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job builder fails to build a pipeline job", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					mock.Anything,
				).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(nil, errors.New("failed"))
				_, err := vaiProvider.CreateRunSchedule(ctx, rsd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job enricher fails", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					mock.Anything,
				).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(nil, errors.New("failed"))
				_, err := vaiProvider.CreateRunSchedule(ctx, rsd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job builder fails to build a schedule", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					mock.Anything,
				).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(&pj, nil)
				mockJobBuilder.On(
					"MkSchedule",
					rsd,
					&pj,
					vaiProvider.config.Parent(),
					vaiProvider.config.GetMaxConcurrentRunCountOrDefault(),
				).Return(nil, errors.New("failed"))
				_, err := vaiProvider.CreateRunSchedule(ctx, rsd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the schedule client fails", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					mock.Anything,
				).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(&pj, nil)
				mockJobBuilder.On(
					"MkSchedule",
					mock.Anything,
					&pj,
					mock.Anything,
					mock.Anything,
				).Return(&aiplatformpb.Schedule{}, nil)
				mockScheduleClient.On("CreateSchedule", mock.Anything).Return(nil, errors.New("failed"))
				_, err := vaiProvider.CreateRunSchedule(ctx, rsd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})

	Context("UpdateRunSchedule", func() {
		When("updating a run schedule", func() {
			It("returns a schedule name", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					fmt.Sprintf(
						"%s/%s/%s",
						rsd.PipelineName.Namespace,
						rsd.PipelineName.Name,
						rsd.PipelineVersion,
					),
				).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(&pj, nil)
				mockJobBuilder.On(
					"MkSchedule",
					rsd,
					&pj,
					vaiProvider.config.Parent(),
					mock.Anything,
				).Return(&aiplatformpb.Schedule{}, nil)
				schedule := aiplatformpb.Schedule{}
				mockScheduleClient.On(
					"UpdateSchedule",
					&aiplatformpb.UpdateScheduleRequest{
						Schedule: &schedule,
						UpdateMask: &fieldmaskpb.FieldMask{
							Paths: []string{
								"schedule",
							},
						},
					},
				).Return(&schedule, nil)
				scheduleName, err := vaiProvider.UpdateRunSchedule(ctx, rsd, "")

				Expect(err).ToNot(HaveOccurred())
				Expect(scheduleName).To(Equal(schedule.Name))
			})

			It("return an error when the file handler read fails", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					mock.Anything,
				).Return(map[string]any{}, errors.New("failed"))
				_, err := vaiProvider.UpdateRunSchedule(ctx, rsd, "")

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job builder fails to build a pipeline job", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					mock.Anything,
				).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(nil, errors.New("failed"))
				_, err := vaiProvider.UpdateRunSchedule(ctx, rsd, "")

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job enricher fails", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					mock.Anything,
				).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(nil, errors.New("failed"))
				_, err := vaiProvider.UpdateRunSchedule(ctx, rsd, "")

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job builder fails to build a schedule", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					mock.Anything,
				).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(&pj, nil)
				mockJobBuilder.On(
					"MkSchedule",
					mock.Anything,
					&pj,
					vaiProvider.config.Parent(),
					mock.Anything,
				).Return(nil, errors.New("failed"))
				_, err := vaiProvider.UpdateRunSchedule(ctx, rsd, "")

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the schedule client fails", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On(
					"Read",
					vaiProvider.config.Parameters.PipelineBucket,
					mock.Anything,
				).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(&pj, nil)
				mockJobBuilder.On(
					"MkSchedule",
					mock.Anything,
					&pj,
					vaiProvider.config.Parent(),
					mock.Anything,
				).Return(nil, errors.New("failed"))
				mockScheduleClient.On("UpdateSchedule", mock.Anything).Return(nil, errors.New("failed"))
				_, err := vaiProvider.UpdateRunSchedule(ctx, rsd, "")

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})

	Context("IsHealthy", func() {
		It("returns true if the provider is healthy", func() {
			mockPipelineClient.On("ListPipelineJobs", mock.Anything, mock.Anything).Return()
			healthy := vaiProvider.IsHealthy(ctx)
			Expect(healthy).To(BeTrue())
		})

		It("returns false if the provider is not healthy", func() {
			mockPipelineClient.On("ListPipelineJobs", mock.Anything, mock.Anything).Return(errors.New("failed"))
			healthy := vaiProvider.IsHealthy(ctx)
			Expect(healthy).To(BeFalse())
		})
	}
})
