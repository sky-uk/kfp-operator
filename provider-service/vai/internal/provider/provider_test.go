//go:build unit

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/config"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/mocks"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type MockJobBuilder struct{ mock.Mock }

func (m *MockJobBuilder) MkRunPipelineJob(
	rd resource.RunDefinition,
) (*aiplatformpb.PipelineJob, error) {
	args := m.Called(rd)
	var pipelineJob *aiplatformpb.PipelineJob
	if arg0 := args.Get(0); arg0 != nil {
		pipelineJob = arg0.(*aiplatformpb.PipelineJob)
	}
	return pipelineJob, args.Error(1)
}

func (m *MockJobBuilder) MkRunSchedulePipelineJob(
	rsd resource.RunScheduleDefinition,
) (*aiplatformpb.PipelineJob, error) {
	args := m.Called(rsd)
	var pipelineJob *aiplatformpb.PipelineJob
	if arg0 := args.Get(0); arg0 != nil {
		pipelineJob = arg0.(*aiplatformpb.PipelineJob)
	}
	return pipelineJob, args.Error(1)
}

func (m *MockJobBuilder) MkSchedule(
	rsd resource.RunScheduleDefinition,
	pipelineJob *aiplatformpb.PipelineJob,
	parent string, maxConcurrentRunCount int64,
) (*aiplatformpb.Schedule, error) {
	args := m.Called(rsd, pipelineJob, parent, maxConcurrentRunCount)
	var schedule *aiplatformpb.Schedule
	if arg0 := args.Get(0); arg0 != nil {
		schedule = arg0.(*aiplatformpb.Schedule)
	}
	return schedule, args.Error(1)
}

type MockJobEnricher struct{ mock.Mock }

func (m *MockJobEnricher) Enrich(
	job *aiplatformpb.PipelineJob,
	raw map[string]any,
) (*aiplatformpb.PipelineJob, error) {
	args := m.Called(job, raw)
	var pipelineJob *aiplatformpb.PipelineJob
	if arg0 := args.Get(0); arg0 != nil {
		pipelineJob = arg0.(*aiplatformpb.PipelineJob)
	}
	return pipelineJob, args.Error(1)
}

// TODO extract to somewhere common
func randomPipelineDefinition() resource.PipelineDefinition {
	return resource.PipelineDefinition{
		Name:          common.RandomNamespacedName(),
		Version:       common.RandomString(),
		Image:         common.RandomString(),
		TfxComponents: common.RandomString(),
		Env:           make([]apis.NamedValue, 0),
		BeamArgs:      make([]apis.NamedValue, 0),
		Manifest:      json.RawMessage{},
	}
}

var _ = Describe("Provider", func() {
	var (
		mockFileHandler    mocks.MockFileHandler
		mockPipelineClient mocks.MockPipelineJobClient
		mockScheduleClient mocks.MockScheduleClient
		mockJobBuilder     MockJobBuilder
		mockJobEnricher    MockJobEnricher
		vaiProvider        VAIProvider
	)

	BeforeEach(func() {
		mockFileHandler = mocks.MockFileHandler{}
		mockPipelineClient = mocks.MockPipelineJobClient{}
		mockScheduleClient = mocks.MockScheduleClient{}
		mockJobBuilder = MockJobBuilder{}
		mockJobEnricher = MockJobEnricher{}
		vaiProvider = VAIProvider{
			ctx:            context.Background(),
			config:         config.VAIProviderConfig{},
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
				mockFileHandler.On("Write", mock.Anything, mock.Anything, mock.Anything).Return(nil)

				pd := randomPipelineDefinition()
				pid, err := vaiProvider.CreatePipeline(pd)

				Expect(err).ToNot(HaveOccurred())
				Expect(pid).To(Equal(fmt.Sprintf("%s/%s", pd.Name.Namespace, pd.Name.Name)))
			})

			It("return an error when the file handler write fails", func() {
				mockFileHandler.On("Write", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed"))

				pd := randomPipelineDefinition()
				_, err := vaiProvider.CreatePipeline(pd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})

	Context("UpdatePipeline", func() {
		When("updating a pipeline", func() {
			It("should return the pipeline ID", func() {
				pd := randomPipelineDefinition()
				mockFileHandler.On(
					"Write",
					mock.MatchedBy(func(j json.RawMessage) bool {
						return bytes.Equal(j, pd.Manifest)
					}),
					mock.MatchedBy(func(s string) bool {
						return s == vaiProvider.config.Parameters.PipelineBucket
					}),
					mock.MatchedBy(func(s string) bool {
						return s == fmt.Sprintf(
							"%s/%s/%s",
							pd.Name.Namespace,
							pd.Name.Name,
							pd.Version,
						)
					}),
				).Return(nil)

				pid, err := vaiProvider.UpdatePipeline(pd, "")

				Expect(err).ToNot(HaveOccurred())
				Expect(pid).To(Equal(fmt.Sprintf(
					"%s/%s", pd.Name.Namespace, pd.Name.Name,
				)))
			})

			It("return an error when the file handler write fails", func() {
				pd := randomPipelineDefinition()
				mockFileHandler.On(
					"Write",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("failed"))

				_, err := vaiProvider.CreatePipeline(pd)

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
					mock.MatchedBy(func(s string) bool {
						return s == id
					}),
					mock.MatchedBy(func(s string) bool {
						return s == vaiProvider.config.Parameters.PipelineBucket
					}),
					mock.Anything,
				).Return(nil)
				err := vaiProvider.DeletePipeline(id)
				Expect(err).ToNot(HaveOccurred())
			})

			It("return an error when the file handler delete fails", func() {
				mockFileHandler.On(
					"Delete",
					"pipelineId",
					mock.Anything,
					mock.Anything,
				).Return(errors.New("failed"))
				err := vaiProvider.DeletePipeline("pipelineId")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})

	Context("CreateRun", func() {
		When("creating a run", func() {
			It("return a run ID", func() {
				rd := randomBasicRunDefinition()
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
					mock.Anything,
					&aiplatformpb.CreatePipelineJobRequest{
						Parent:        vaiProvider.config.Parent(),
						PipelineJobId: fmt.Sprintf("%s-%s-%s", rd.Name.Namespace, rd.Name.Name, rd.Version),
						PipelineJob:   &pj,
					},
					mock.Anything,
				).Return(&pj, nil)
				runId, err := vaiProvider.CreateRun(rd)

				Expect(err).ToNot(HaveOccurred())
				Expect(runId).To(Equal(fmt.Sprintf("%s-%s", rd.Name.Namespace, rd.Name.Name)))
			})

			It("return an error when the file handler read fails", func() {
				rd := randomBasicRunDefinition()
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, errors.New("failed"))
				_, err := vaiProvider.CreateRun(rd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job builder fails", func() {
				rd := randomBasicRunDefinition()
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunPipelineJob", rd).Return(nil, errors.New("failed"))
				_, err := vaiProvider.CreateRun(rd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job enricher fails", func() {
				rd := randomBasicRunDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunPipelineJob", rd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(nil, errors.New("failed"))
				_, err := vaiProvider.CreateRun(rd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the pipeline client fails", func() {
				rd := randomBasicRunDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunPipelineJob", rd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(&pj, nil)
				mockPipelineClient.On("CreatePipelineJob", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				_, err := vaiProvider.CreateRun(rd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})

	Context("CreateRunSchedule", func() {
		When("creating a run schedule", func() {
			It("returns a schedule name", func() {
				rsd := randomRunScheduleDefinition()
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
					mock.Anything,
					&aiplatformpb.CreateScheduleRequest{
						Parent:   vaiProvider.config.Parent(),
						Schedule: &schedule,
					},
					mock.Anything,
				).Return(&schedule, nil)
				scheduleName, err := vaiProvider.CreateRunSchedule(rsd)

				Expect(err).ToNot(HaveOccurred())
				Expect(scheduleName).To(Equal(schedule.Name))
			})

			It("return an error when the file handler read fails", func() {
				rsd := randomRunScheduleDefinition()
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, errors.New("failed"))
				_, err := vaiProvider.CreateRunSchedule(rsd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job builder fails to build a pipeline job", func() {
				rsd := randomRunScheduleDefinition()
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(nil, errors.New("failed"))
				_, err := vaiProvider.CreateRunSchedule(rsd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job enricher fails", func() {
				rsd := randomRunScheduleDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(nil, errors.New("failed"))
				_, err := vaiProvider.CreateRunSchedule(rsd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job builder fails to build a schedule", func() {
				rsd := randomRunScheduleDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(&pj, nil)
				mockJobBuilder.On("MkSchedule", mock.Anything, &pj, mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				_, err := vaiProvider.CreateRunSchedule(rsd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the schedule client fails", func() {
				rsd := randomRunScheduleDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(&pj, nil)
				mockJobBuilder.On("MkSchedule", mock.Anything, &pj, mock.Anything, mock.Anything).Return(&aiplatformpb.Schedule{}, nil)
				mockScheduleClient.On("CreateSchedule", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				_, err := vaiProvider.CreateRunSchedule(rsd)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})

	Context("UpdateRunSchedule", func() {
		When("updating a run schedule", func() {
			It("returns a schedule name", func() {
				rsd := randomRunScheduleDefinition()
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
					mock.Anything,
					&aiplatformpb.UpdateScheduleRequest{
						Schedule: &schedule,
						UpdateMask: &fieldmaskpb.FieldMask{
							Paths: []string{
								"schedule",
							},
						},
					},
					mock.Anything,
				).Return(&schedule, nil)
				scheduleName, err := vaiProvider.UpdateRunSchedule(rsd, "")

				Expect(err).ToNot(HaveOccurred())
				Expect(scheduleName).To(Equal(schedule.Name))
			})

			It("return an error when the file handler read fails", func() {
				rsd := randomRunScheduleDefinition()
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, errors.New("failed"))
				_, err := vaiProvider.UpdateRunSchedule(rsd, "")

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job builder fails to build a pipeline job", func() {
				rsd := randomRunScheduleDefinition()
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(nil, errors.New("failed"))
				_, err := vaiProvider.UpdateRunSchedule(rsd, "")

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job enricher fails", func() {
				rsd := randomRunScheduleDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(nil, errors.New("failed"))
				_, err := vaiProvider.UpdateRunSchedule(rsd, "")

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job builder fails to build a schedule", func() {
				rsd := randomRunScheduleDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(&pj, nil)
				mockJobBuilder.On("MkSchedule", mock.Anything, &pj, mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				_, err := vaiProvider.UpdateRunSchedule(rsd, "")

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the schedule client fails", func() {
				rsd := randomRunScheduleDefinition()
				pj := aiplatformpb.PipelineJob{}
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", rsd).Return(&pj, nil)
				mockJobEnricher.On("Enrich", &pj, map[string]any{}).Return(&pj, nil)
				mockJobBuilder.On("MkSchedule", mock.Anything, &pj, mock.Anything, mock.Anything).Return(&aiplatformpb.Schedule{}, nil)
				mockScheduleClient.On("UpdateSchedule", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				_, err := vaiProvider.UpdateRunSchedule(rsd, "")

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})
})
