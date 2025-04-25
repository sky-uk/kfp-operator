//go:build unit

package provider

import (
	"context"
	"errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/testutil"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Provider", func() {
	var (
		provider                  *KfpProvider
		mockPipelineService       mocks.MockPipelineService
		mockPipelineUploadService mocks.MockPipelineUploadService
		mockRunService            mocks.MockRunService
		mockExperimentService     mocks.MockExperimentService
		mockJobService            mocks.MockJobService
		ctx                       = context.Background()
	)

	const ignoreCtx = mock.Anything

	BeforeEach(func() {
		mockPipelineService = mocks.MockPipelineService{}
		mockPipelineUploadService = mocks.MockPipelineUploadService{}
		mockRunService = mocks.MockRunService{}
		mockExperimentService = mocks.MockExperimentService{}
		mockJobService = mocks.MockJobService{}

		provider = &KfpProvider{
			config:                &config.KfpProviderConfig{},
			pipelineUploadService: &mockPipelineUploadService,
			pipelineService:       &mockPipelineService,
			runService:            &mockRunService,
			experimentService:     &mockExperimentService,
			jobService:            &mockJobService,
		}
	})

	Context("Run", func() {
		Context("CreateRun", func() {
			const (
				runId             = "run-id"
				pipelineId        = "pipeline-id"
				pipelineVersionId = "pipeline-version-id"
				experimentId      = "experiment-id"
			)
			rd := testutil.RandomRunDefinition()
			nsnStr, err := rd.PipelineName.SeparatedString("-")

			Expect(err).ToNot(HaveOccurred())

			It("should return run id if run is created", func() {
				mockPipelineService.On("PipelineIdForName", ignoreCtx, nsnStr).Return(pipelineId, nil)
				mockPipelineService.On("PipelineVersionIdForName", ignoreCtx, rd.PipelineVersion, pipelineId).Return(pipelineVersionId, nil)
				mockExperimentService.On("ExperimentIdByName", ignoreCtx, rd.ExperimentName).Return(experimentId, nil)
				mockRunService.On("CreateRun", ignoreCtx, rd, pipelineId, pipelineVersionId, experimentId).Return(runId, nil)
				result, err := provider.CreateRun(ctx, rd)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(runId))
			})

			When("pipeline has invalid namespace name", func() {
				It("should return error", func() {
					copyRd := rd
					copyRd.PipelineName.Name = ""
					result, err := provider.CreateRun(ctx, copyRd)

					Expect(err).To(HaveOccurred())
					Expect(result).To(BeEmpty())
				})
			})

			When("pipeline service PipelineIdForName errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					mockPipelineService.On("PipelineIdForName", ignoreCtx, nsnStr).Return("", expectedErr)
					result, err := provider.CreateRun(ctx, rd)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})

			When("pipeline service PipelineVersionIdForName errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					mockPipelineService.On("PipelineIdForName", ignoreCtx, nsnStr).Return(pipelineId, nil)
					mockPipelineService.On("PipelineVersionIdForName", ignoreCtx, rd.PipelineVersion, pipelineId).Return("", expectedErr)
					result, err := provider.CreateRun(ctx, rd)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})

			When("experiment service ExperimentIdByName errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					mockPipelineService.On("PipelineIdForName", ignoreCtx, nsnStr).Return(pipelineId, nil)
					mockPipelineService.On("PipelineVersionIdForName", ignoreCtx, rd.PipelineVersion, pipelineId).Return(pipelineVersionId, nil)
					mockExperimentService.On("ExperimentIdByName", ignoreCtx, rd.ExperimentName).Return("", expectedErr)
					result, err := provider.CreateRun(ctx, rd)

					Expect(err).To(HaveOccurred())
					Expect(result).To(BeEmpty())
				})
			})

			When("run service CreateRun errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					mockPipelineService.On("PipelineIdForName", ignoreCtx, nsnStr).Return(pipelineId, nil)
					mockPipelineService.On("PipelineVersionIdForName", ignoreCtx, rd.PipelineVersion, pipelineId).Return(pipelineVersionId, nil)
					mockExperimentService.On("ExperimentIdByName", ignoreCtx, rd.ExperimentName).Return(experimentId, nil)
					mockRunService.On("CreateRun", ignoreCtx, rd, pipelineId, pipelineVersionId, experimentId).Return("", expectedErr)
					result, err := provider.CreateRun(ctx, rd)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})
		})
	})

	Context("Pipeline", func() {
		const id = "pipeline-id"
		pdw := testutil.RandomPipelineDefinitionWrapper()
		version := pdw.PipelineDefinition.Version
		nsnStr, err := pdw.PipelineDefinition.Name.String()

		Expect(err).ToNot(HaveOccurred())

		Context("CreatePipeline", func() {
			It("should return id if pipeline is created", func() {
				mockPipelineUploadService.On("UploadPipeline", ignoreCtx, []byte{}, nsnStr).Return(id, nil)
				mockPipelineUploadService.On("UploadPipelineVersion", ignoreCtx, id, []byte{}, version).Return(nil)
				result, err := provider.CreatePipeline(ctx, pdw)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(id))
			})

			It("should return err if namespace / name is invalid", func() {
				copyPipeline := pdw
				copyPipeline.PipelineDefinition.Name.Name = ""
				result, err := provider.CreatePipeline(ctx, copyPipeline)

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeEmpty())
			})

			It("should return err if UploadPipeline fails", func() {
				expectedErr := errors.New("failed")
				mockPipelineUploadService.On("UploadPipeline", ignoreCtx, []byte{}, nsnStr).Return("", expectedErr)
				result, err := provider.CreatePipeline(ctx, pdw)

				Expect(err).To(Equal(expectedErr))
				Expect(result).To(BeEmpty())
			})

			It("should return err if UpdatePipelineVersion fails", func() {
				expectedErr := errors.New("failed")
				mockPipelineUploadService.On("UploadPipeline", ignoreCtx, []byte{}, nsnStr).Return(id, nil)
				mockPipelineUploadService.On("UploadPipelineVersion", ignoreCtx, id, []byte{}, version).Return(expectedErr)
				result, err := provider.CreatePipeline(ctx, pdw)

				Expect(err).To(Equal(expectedErr))
				Expect(result).To(BeEmpty())
			})
		})

		Context("UpdatePipeline", func() {
			It("should return id if pipeline is updated", func() {
				mockPipelineUploadService.On("UploadPipelineVersion", ignoreCtx, id, []byte{}, version).Return(nil)
				result, err := provider.UpdatePipeline(ctx, pdw, id)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(id))
			})

			When("pipeline upload service errors", func() {
				It("should return empty id and err", func() {
					expectedErr := errors.New("failed")
					mockPipelineUploadService.On("UploadPipelineVersion", ignoreCtx, id, []byte{}, version).Return(expectedErr)
					result, err := provider.UpdatePipeline(ctx, pdw, id)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})

		})

		Context("DeletePipeline", func() {
			It("should not error", func() {
				mockPipelineService.On("DeletePipeline", ignoreCtx, id).Return(nil)

				Expect(provider.DeletePipeline(ctx, id)).To(Succeed())
			})

			When("pipeline service errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					mockPipelineService.On("DeletePipeline", ignoreCtx, id).Return(expectedErr)
					err := provider.DeletePipeline(ctx, id)

					Expect(err).To(Equal(expectedErr))
				})
			})
		})
	})

	Context("RunSchedule", func() {
		const (
			jobId             = "job-id"
			pipelineId        = "pipeline-id"
			pipelineVersionId = "pipeline-version-id"
			experimentId      = "experiment-id"
		)
		rsd := testutil.RandomRunScheduleDefinition()
		nsnStr, err := rsd.PipelineName.SeparatedString("-")

		Expect(err).ToNot(HaveOccurred())

		Context("CreateRunSchedule", func() {
			It("should return job id if run schedule is created", func() {
				mockPipelineService.On("PipelineIdForName", ignoreCtx, nsnStr).Return(pipelineId, nil)
				mockPipelineService.On("PipelineVersionIdForName", ignoreCtx, rsd.PipelineVersion, pipelineId).Return(pipelineVersionId, nil)
				mockExperimentService.On("ExperimentIdByName", ignoreCtx, rsd.ExperimentName).Return(experimentId, nil)
				mockJobService.On("CreateJob", ignoreCtx, rsd, pipelineId, pipelineVersionId, experimentId).Return(jobId, nil)
				result, err := provider.CreateRunSchedule(ctx, rsd)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(jobId))
			})

			When("pipeline has invalid namespace name", func() {
				It("should return error", func() {
					copyRsd := rsd
					copyRsd.PipelineName.Name = ""
					result, err := provider.CreateRunSchedule(ctx, copyRsd)

					Expect(err).To(HaveOccurred())
					Expect(result).To(BeEmpty())
				})
			})

			When("pipeline service PipelineIdForName errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					mockPipelineService.On("PipelineIdForName", ignoreCtx, nsnStr).Return("", expectedErr)
					result, err := provider.CreateRunSchedule(ctx, rsd)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})

			When("pipeline service PipelineVersionIdForName errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					mockPipelineService.On("PipelineIdForName", ignoreCtx, nsnStr).Return(pipelineId, nil)
					mockPipelineService.On("PipelineVersionIdForName", ignoreCtx, rsd.PipelineVersion, pipelineId).Return("", expectedErr)
					result, err := provider.CreateRunSchedule(ctx, rsd)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})

			When("experiment service ExperimentIdByName errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					mockPipelineService.On("PipelineIdForName", ignoreCtx, nsnStr).Return(pipelineId, nil)
					mockPipelineService.On("PipelineVersionIdForName", ignoreCtx, rsd.PipelineVersion, pipelineId).Return(pipelineVersionId, nil)
					mockExperimentService.On("ExperimentIdByName", ignoreCtx, rsd.ExperimentName).Return("", expectedErr)
					result, err := provider.CreateRunSchedule(ctx, rsd)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})

			When("job service CreateJob errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					mockPipelineService.On("PipelineIdForName", ignoreCtx, nsnStr).Return(pipelineId, nil)
					mockPipelineService.On("PipelineVersionIdForName", ignoreCtx, rsd.PipelineVersion, pipelineId).Return(pipelineVersionId, nil)
					mockExperimentService.On("ExperimentIdByName", ignoreCtx, rsd.ExperimentName).Return(experimentId, nil)
					mockJobService.On("CreateJob", ignoreCtx, rsd, pipelineId, pipelineVersionId, experimentId).Return("", expectedErr)
					result, err := provider.CreateRunSchedule(ctx, rsd)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})
		})

		Context("UpdateRunSchedule", func() {
			It("should return job id if run schedule is updated", func() {
				mockJobService.On("DeleteJob", ignoreCtx, jobId).Return(nil)
				mockPipelineService.On("PipelineIdForName", ignoreCtx, nsnStr).Return(pipelineId, nil)
				mockPipelineService.On("PipelineVersionIdForName", ignoreCtx, rsd.PipelineVersion, pipelineId).Return(pipelineVersionId, nil)
				mockExperimentService.On("ExperimentIdByName", ignoreCtx, rsd.ExperimentName).Return(experimentId, nil)
				mockJobService.On("CreateJob", ignoreCtx, rsd, pipelineId, pipelineVersionId, experimentId).Return(jobId, nil)
				result, err := provider.UpdateRunSchedule(ctx, rsd, jobId)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(jobId))
			})

			When("DeleteRunSchedule errors", func() {
				It("should return error and retain the id", func() {
					expectedErr := errors.New("failed")
					mockJobService.On("DeleteJob", ignoreCtx, jobId).Return(expectedErr)
					result, err := provider.UpdateRunSchedule(ctx, rsd, jobId)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(Equal(jobId))
				})
			})

			When("CreateRunSchedule errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					mockJobService.On("DeleteJob", ignoreCtx, jobId).Return(nil)
					mockPipelineService.On("PipelineIdForName", ignoreCtx, nsnStr).Return("", expectedErr)
					result, err := provider.UpdateRunSchedule(ctx, rsd, jobId)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})
		})

		Context("DeleteRunSchedule", func() {
			It("should not error", func() {
				mockJobService.On("DeleteJob", ignoreCtx, jobId).Return(nil)

				Expect(provider.DeleteRunSchedule(ctx, jobId)).To(Succeed())
			})

			When("job service errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					mockJobService.On("DeleteJob", ignoreCtx, jobId).Return(expectedErr)
					err := provider.DeleteRunSchedule(ctx, jobId)

					Expect(err).To(Equal(expectedErr))
				})
			})
		})
	})

	Context("Experiment", func() {
		const id = "experiment-id"
		experiment := testutil.RandomExperimentDefinition()

		Context("CreateExperiment", func() {
			It("should return id if experiment is created", func() {
				mockExperimentService.On("CreateExperiment", ignoreCtx, experiment.Name, experiment.Description).Return(id, nil)
				result, err := provider.CreateExperiment(ctx, experiment)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(id))
			})

			When("experiment service CreateExperiment errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					mockExperimentService.On("CreateExperiment", ignoreCtx, experiment.Name, experiment.Description).Return("", expectedErr)
					result, err := provider.CreateExperiment(ctx, experiment)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})
		})

		Context("UpdateExperiment", func() {
			It("should return id if experiment is updated", func() {
				mockExperimentService.On("DeleteExperiment", ignoreCtx, id).Return(nil)
				mockExperimentService.On("CreateExperiment", ignoreCtx, experiment.Name, experiment.Description).Return(id, nil)
				result, err := provider.UpdateExperiment(ctx, experiment, id)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(id))
			})

			When("DeleteExperiment errors", func() {
				It("should return error and retain the id", func() {
					expectedErr := errors.New("failed")
					mockExperimentService.On("DeleteExperiment", ignoreCtx, id).Return(expectedErr)
					result, err := provider.UpdateExperiment(ctx, experiment, id)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(Equal(id))
				})
			})

			When("CreateExperiment errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					mockExperimentService.On("DeleteExperiment", ignoreCtx, id).Return(nil)
					mockExperimentService.On("CreateExperiment", ignoreCtx, experiment.Name, experiment.Description).Return("", expectedErr)
					result, err := provider.UpdateExperiment(ctx, experiment, id)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})
		})

		Context("DeleteExperiment", func() {
			It("should not error", func() {
				mockExperimentService.On("DeleteExperiment", ignoreCtx, id).Return(nil)

				Expect(provider.DeleteExperiment(ctx, id)).To(Succeed())
			})

			When("experiment service errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					mockExperimentService.On("DeleteExperiment", ignoreCtx, id).Return(expectedErr)
					err := provider.DeleteExperiment(ctx, id)

					Expect(err).To(Equal(expectedErr))
				})
			})
		})
	})
})
