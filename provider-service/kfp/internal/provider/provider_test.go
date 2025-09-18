//go:build unit

package provider

import (
	"context"
	"errors"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/label"
	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/testutil"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/util"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
)

var _ = Describe("Provider", func() {
	var (
		provider              *KfpProvider
		pipelineService       mocks.MockPipelineService
		pipelineUploadService mocks.MockPipelineUploadService
		runService            mocks.MockRunService
		experimentService     mocks.MockExperimentService
		recurringRunService   mocks.MockRecurringRunService
		labelService          mocks.MockLabelService
		ctx                   = context.Background()
	)

	BeforeEach(func() {
		pipelineService = mocks.MockPipelineService{}
		pipelineUploadService = mocks.MockPipelineUploadService{}
		runService = mocks.MockRunService{}
		experimentService = mocks.MockExperimentService{}
		recurringRunService = mocks.MockRecurringRunService{}
		labelService = mocks.MockLabelService{}

		provider = &KfpProvider{
			config:                &config.Config{},
			pipelineUploadService: &pipelineUploadService,
			pipelineService:       &pipelineService,
			runService:            &runService,
			experimentService:     &experimentService,
			recurringRunService:   &recurringRunService,
			labelService:          &labelService,
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
				pipelineService.On("PipelineIdForDisplayName", nsnStr).Return(pipelineId, nil)
				pipelineService.On("PipelineVersionIdForDisplayName", rd.PipelineVersion, pipelineId).Return(pipelineVersionId, nil)
				experimentService.On("ExperimentIdByDisplayName", rd.ExperimentName).Return(experimentId, nil)
				runService.On("CreateRun", rd, pipelineId, pipelineVersionId, experimentId).Return(runId, nil)
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

			When("pipeline service PipelineIdForDisplayName errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					pipelineService.On("PipelineIdForDisplayName", nsnStr).Return("", expectedErr)
					result, err := provider.CreateRun(ctx, rd)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})

			When("pipeline service PipelineVersionIdForDisplayName errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					pipelineService.On("PipelineIdForDisplayName", nsnStr).Return(pipelineId, nil)
					pipelineService.On("PipelineVersionIdForDisplayName", rd.PipelineVersion, pipelineId).Return("", expectedErr)
					result, err := provider.CreateRun(ctx, rd)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})

			When("experiment service ExperimentIdByDisplayName errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					pipelineService.On("PipelineIdForDisplayName", nsnStr).Return(pipelineId, nil)
					pipelineService.On("PipelineVersionIdForDisplayName", rd.PipelineVersion, pipelineId).Return(pipelineVersionId, nil)
					experimentService.On("ExperimentIdByDisplayName", rd.ExperimentName).Return("", expectedErr)
					result, err := provider.CreateRun(ctx, rd)

					Expect(err).To(HaveOccurred())
					Expect(result).To(BeEmpty())
				})
			})

			When("run service CreateRun errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					pipelineService.On("PipelineIdForDisplayName", nsnStr).Return(pipelineId, nil)
					pipelineService.On("PipelineVersionIdForDisplayName", rd.PipelineVersion, pipelineId).Return(pipelineVersionId, nil)
					experimentService.On("ExperimentIdByDisplayName", rd.ExperimentName).Return(experimentId, nil)
					runService.On("CreateRun", rd, pipelineId, pipelineVersionId, experimentId).Return("", expectedErr)
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
		nsnStr, err := util.ResourceNameFromNamespacedName(pdw.PipelineDefinition.Name)

		Expect(err).ToNot(HaveOccurred())

		Context("CreatePipeline", func() {
			It("should return id if pipeline is created", func() {
				pipelineUploadService.On("UploadPipeline", mock.Anything, nsnStr).Return(id, nil)
				pipelineService.On("DeletePipelineVersions", id).Return(nil)
				pipelineUploadService.On("UploadPipelineVersion", id, mock.Anything, version).Return(nil)
				labelService.On("InsertLabelsIntoParameters", mock.Anything, label.LabelKeys).Return(pdw.CompiledPipeline, nil)
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
				labelService.On("InsertLabelsIntoParameters", mock.Anything, label.LabelKeys).Return(pdw.CompiledPipeline, nil)
				pipelineUploadService.On("UploadPipeline", mock.Anything, nsnStr).Return("", expectedErr)
				result, err := provider.CreatePipeline(ctx, pdw)

				Expect(err).To(Equal(expectedErr))
				Expect(result).To(BeEmpty())
			})

			It("should return err if DeletePipelineVersions fails", func() {
				expectedErr := errors.New("failed")
				labelService.On("InsertLabelsIntoParameters", mock.Anything, label.LabelKeys).Return(pdw.CompiledPipeline, nil)
				pipelineUploadService.On("UploadPipeline", mock.Anything, nsnStr).Return(id, nil)
				pipelineService.On("DeletePipelineVersions", id).Return(expectedErr)
				result, err := provider.CreatePipeline(ctx, pdw)

				Expect(err).To(Equal(expectedErr))
				Expect(result).To(BeEmpty())
			})

			It("should return err if UploadPipelineVersion fails", func() {
				expectedErr := errors.New("failed")
				labelService.On("InsertLabelsIntoParameters", mock.Anything, label.LabelKeys).Return(pdw.CompiledPipeline, nil)
				pipelineUploadService.On("UploadPipeline", mock.Anything, nsnStr).Return(id, nil)
				pipelineService.On("DeletePipelineVersions", id).Return(nil)
				pipelineUploadService.On("UploadPipelineVersion", id, mock.Anything, version).Return(expectedErr)
				result, err := provider.CreatePipeline(ctx, pdw)

				Expect(err).To(Equal(expectedErr))
				Expect(result).To(BeEmpty())
			})
		})

		Context("UpdatePipeline", func() {
			It("should return id if pipeline versions are cleaned up and version is updated", func() {
				pipelineService.On("DeletePipelineVersions", id).Return(nil)
				pipelineUploadService.On("UploadPipelineVersion", id, mock.Anything, version).Return(nil)
				labelService.On("InsertLabelsIntoParameters", mock.Anything, label.LabelKeys).Return(pdw.CompiledPipeline, nil)
				result, err := provider.UpdatePipeline(ctx, pdw, id)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(id))
			})

			When("DeletePipelineVersions errors", func() {
				It("should return empty id and err", func() {
					expectedErr := errors.New("failed")
					pipelineService.On("DeletePipelineVersions", id).Return(expectedErr)
					result, err := provider.UpdatePipeline(ctx, pdw, id)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})

			When("UploadPipelineVersion errors", func() {
				It("should return empty id and err", func() {
					expectedErr := errors.New("failed")
					pipelineService.On("DeletePipelineVersions", id).Return(nil)
					pipelineUploadService.On("UploadPipelineVersion", id, mock.Anything, version).Return(expectedErr)
					labelService.On("InsertLabelsIntoParameters", mock.Anything, label.LabelKeys).Return(pdw.CompiledPipeline, nil)

					result, err := provider.UpdatePipeline(ctx, pdw, id)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})

		})

		Context("DeletePipeline", func() {
			It("should not error", func() {
				pipelineService.On("DeletePipelineVersions", id).Return(nil)
				pipelineService.On("DeletePipeline", id).Return(nil)

				Expect(provider.DeletePipeline(ctx, id)).To(Succeed())
			})

			When("DeletePipeline errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					pipelineService.On("DeletePipelineVersions", id).Return(nil)
					pipelineService.On("DeletePipeline", id).Return(expectedErr)
					err := provider.DeletePipeline(ctx, id)

					Expect(err).To(Equal(expectedErr))
				})
			})

			When("pipeline service DeletePipelineVersions errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					pipelineService.On("DeletePipelineVersions", id).Return(expectedErr)
					err := provider.DeletePipeline(ctx, id)

					Expect(err).To(Equal(expectedErr))
				})
			})

			When("pipeline service DeletePipeline errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					pipelineService.On("DeletePipelineVersions", id).Return(nil)
					pipelineService.On("DeletePipeline", id).Return(expectedErr)
					err := provider.DeletePipeline(ctx, id)

					Expect(err).To(Equal(expectedErr))
				})
			})
		})
	})

	Context("RunSchedule", func() {
		const (
			recurringRunId    = "recurring-run-id"
			pipelineId        = "pipeline-id"
			pipelineVersionId = "pipeline-version-id"
			experimentId      = "experiment-id"
		)
		rsd := testutil.RandomRunScheduleDefinition()
		nsnStr, err := rsd.PipelineName.SeparatedString("-")

		Expect(err).ToNot(HaveOccurred())

		Context("CreateRunSchedule", func() {
			It("should return recurring run id if run schedule is created", func() {
				pipelineService.On("PipelineIdForDisplayName", nsnStr).Return(pipelineId, nil)
				pipelineService.On("PipelineVersionIdForDisplayName", rsd.PipelineVersion, pipelineId).Return(pipelineVersionId, nil)
				experimentService.On("ExperimentIdByDisplayName", rsd.ExperimentName).Return(experimentId, nil)
				recurringRunService.On("CreateRecurringRun", rsd, pipelineId, pipelineVersionId, experimentId).Return(recurringRunId, nil)
				result, err := provider.CreateRunSchedule(ctx, rsd)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(recurringRunId))
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

			When("pipeline service PipelineIdForDisplayName errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					pipelineService.On("PipelineIdForDisplayName", nsnStr).Return("", expectedErr)
					result, err := provider.CreateRunSchedule(ctx, rsd)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})

			When("pipeline service PipelineVersionIdForDisplayName errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					pipelineService.On("PipelineIdForDisplayName", nsnStr).Return(pipelineId, nil)
					pipelineService.On("PipelineVersionIdForDisplayName", rsd.PipelineVersion, pipelineId).Return("", expectedErr)
					result, err := provider.CreateRunSchedule(ctx, rsd)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})

			When("experiment service ExperimentIdByDisplayName errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					pipelineService.On("PipelineIdForDisplayName", nsnStr).Return(pipelineId, nil)
					pipelineService.On("PipelineVersionIdForDisplayName", rsd.PipelineVersion, pipelineId).Return(pipelineVersionId, nil)
					experimentService.On("ExperimentIdByDisplayName", rsd.ExperimentName).Return("", expectedErr)
					result, err := provider.CreateRunSchedule(ctx, rsd)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})

			When("recurring run service CreateRecurringRun errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					pipelineService.On("PipelineIdForDisplayName", nsnStr).Return(pipelineId, nil)
					pipelineService.On("PipelineVersionIdForDisplayName", rsd.PipelineVersion, pipelineId).Return(pipelineVersionId, nil)
					experimentService.On("ExperimentIdByDisplayName", rsd.ExperimentName).Return(experimentId, nil)
					recurringRunService.On("CreateRecurringRun", rsd, pipelineId, pipelineVersionId, experimentId).Return("", expectedErr)
					result, err := provider.CreateRunSchedule(ctx, rsd)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})
		})

		Context("UpdateRunSchedule", func() {
			It("should return recurring run id if run schedule is updated", func() {
				recurringRunService.On("DeleteRecurringRun", recurringRunId).Return(nil)
				pipelineService.On("PipelineIdForDisplayName", nsnStr).Return(pipelineId, nil)
				pipelineService.On("PipelineVersionIdForDisplayName", rsd.PipelineVersion, pipelineId).Return(pipelineVersionId, nil)
				experimentService.On("ExperimentIdByDisplayName", rsd.ExperimentName).Return(experimentId, nil)
				recurringRunService.On("CreateRecurringRun", rsd, pipelineId, pipelineVersionId, experimentId).Return(recurringRunId, nil)
				result, err := provider.UpdateRunSchedule(ctx, rsd, recurringRunId)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(recurringRunId))
			})

			When("DeleteRunSchedule errors", func() {
				It("should return error and retain the id", func() {
					expectedErr := errors.New("failed")
					recurringRunService.On("DeleteRecurringRun", recurringRunId).Return(expectedErr)
					result, err := provider.UpdateRunSchedule(ctx, rsd, recurringRunId)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(Equal(recurringRunId))
				})
			})

			When("CreateRunSchedule errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					recurringRunService.On("DeleteRecurringRun", recurringRunId).Return(nil)
					pipelineService.On("PipelineIdForDisplayName", nsnStr).Return("", expectedErr)
					result, err := provider.UpdateRunSchedule(ctx, rsd, recurringRunId)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})
		})

		Context("DeleteRunSchedule", func() {
			It("should not error", func() {
				recurringRunService.On("DeleteRecurringRun", recurringRunId).Return(nil)

				Expect(provider.DeleteRunSchedule(ctx, recurringRunId)).To(Succeed())
			})

			When("recurring run service errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					recurringRunService.On("DeleteRecurringRun", recurringRunId).Return(expectedErr)
					err := provider.DeleteRunSchedule(ctx, recurringRunId)

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
				experimentService.On("CreateExperiment", experiment.Name, experiment.Description).Return(id, nil)
				result, err := provider.CreateExperiment(ctx, experiment)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(id))
			})

			When("experiment service CreateExperiment errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					experimentService.On("CreateExperiment", experiment.Name, experiment.Description).Return("", expectedErr)
					result, err := provider.CreateExperiment(ctx, experiment)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})
		})

		Context("UpdateExperiment", func() {
			It("should return id if experiment is updated", func() {
				experimentService.On("DeleteExperiment", id).Return(nil)
				experimentService.On("CreateExperiment", experiment.Name, experiment.Description).Return(id, nil)
				result, err := provider.UpdateExperiment(ctx, experiment, id)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(id))
			})

			When("DeleteExperiment errors", func() {
				It("should return error and retain the id", func() {
					expectedErr := errors.New("failed")
					experimentService.On("DeleteExperiment", id).Return(expectedErr)
					result, err := provider.UpdateExperiment(ctx, experiment, id)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(Equal(id))
				})
			})

			When("CreateExperiment errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					experimentService.On("DeleteExperiment", id).Return(nil)
					experimentService.On("CreateExperiment", experiment.Name, experiment.Description).Return("", expectedErr)
					result, err := provider.UpdateExperiment(ctx, experiment, id)

					Expect(err).To(Equal(expectedErr))
					Expect(result).To(BeEmpty())
				})
			})
		})

		Context("DeleteExperiment", func() {
			It("should not error", func() {
				experimentService.On("DeleteExperiment", id).Return(nil)

				Expect(provider.DeleteExperiment(ctx, id)).To(Succeed())
			})

			When("experiment service errors", func() {
				It("should return error", func() {
					expectedErr := errors.New("failed")
					experimentService.On("DeleteExperiment", id).Return(expectedErr)
					err := provider.DeleteExperiment(ctx, id)

					Expect(err).To(Equal(expectedErr))
				})
			})
		})
	})
})
