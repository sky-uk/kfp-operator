//go:build unit

package vai

import (
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"context"
	"errors"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/argo/providers/base"
)

func randomBasicRunIntent() RunIntent {
	return RunIntent{
		PipelineName:    common.RandomNamespacedName(),
		PipelineVersion: common.RandomString(),
	}
}

func randomRunScheduleDefinition() RunScheduleDefinition {
	return RunScheduleDefinition{
		Name:                 common.RandomString(),
		Version:              common.RandomString(),
		PipelineName:         common.RandomNamespacedName(),
		PipelineVersion:      common.RandomString(),
		RunConfigurationName: common.RandomNamespacedName(),
		ExperimentName:       common.RandomString(),
		Schedule:             "1 1 0 0 0",
	}
}

func randomVAIProviderConfig() VAIProviderConfig {
	return VAIProviderConfig{
		PipelineBucket: common.RandomString(),
		VaiProject:     common.RandomString(),
		VaiLocation:    common.RandomString(),
	}
}

var _ = Context("VAI Provider", func() {
	ctx := context.Background()

	Describe("runLabelsFromSchedule", func() {
		It("generates run labels without RunConfigurationName or RunName", func() {
			input := randomRunScheduleDefinition()
			input.RunConfigurationName = common.NamespacedName{}

			runLabels := runLabelsFromSchedule(input)

			expectedRunLabels := map[string]string{
				labels.PipelineName:      input.PipelineName.Name,
				labels.PipelineNamespace: input.PipelineName.Namespace,
				labels.PipelineVersion:   input.PipelineVersion,
			}

			Expect(runLabels).To(Equal(expectedRunLabels))
		})

		It("generates run labels with RunConfigurationName", func() {
			input := randomRunScheduleDefinition()

			runLabels := runLabelsFromSchedule(input)

			Expect(runLabels[labels.RunConfigurationName]).To(Equal(input.RunConfigurationName.Name))
			Expect(runLabels[labels.RunConfigurationNamespace]).To(Equal(input.RunConfigurationName.Namespace))
			Expect(runLabels).NotTo(HaveKey(labels.RunName))
			Expect(runLabels).NotTo(HaveKey(labels.RunNamespace))
		})

		It("replaces fullstops with dashes in pipelineVersion", func() {
			input := randomBasicRunIntent()
			input.PipelineVersion = "0.4.0"

			runLabels := runLabelsFromRun(input)

			Expect(runLabels[labels.PipelineVersion]).To(Equal("0-4-0"))
		})
	})

	Describe("runLabelsFromRun", func() {
		It("generates run labels with RunName", func() {
			input := randomBasicRunIntent()
			input.RunName = common.RandomNamespacedName()

			runLabels := runLabelsFromRun(input)

			Expect(runLabels[labels.RunName]).To(Equal(input.RunName.Name))
			Expect(runLabels[labels.RunNamespace]).To(Equal(input.RunName.Namespace))
			Expect(runLabels).NotTo(HaveKey(labels.RunConfigurationName))
			Expect(runLabels).NotTo(HaveKey(labels.RunConfigurationNamespace))
		})

		It("generates run labels with RunConfigurationName and RunName", func() {
			input := randomBasicRunIntent()
			input.RunConfigurationName = common.RandomNamespacedName()
			input.RunName = common.RandomNamespacedName()

			runLabels := runLabelsFromRun(input)

			Expect(runLabels[labels.RunConfigurationName]).To(Equal(input.RunConfigurationName.Name))
			Expect(runLabels[labels.RunConfigurationNamespace]).To(Equal(input.RunConfigurationName.Namespace))
			Expect(runLabels[labels.RunName]).To(Equal(input.RunName.Name))
			Expect(runLabels[labels.RunNamespace]).To(Equal(input.RunName.Namespace))
		})

		It("replaces fullstops with dashes in pipelineVersion", func() {
			input := randomBasicRunIntent()
			input.PipelineVersion = "0.4.0"

			runLabels := runLabelsFromRun(input)

			Expect(runLabels[labels.PipelineVersion]).To(Equal("0-4-0"))
		})
	})

	Describe("isLegacySchedule", func() {
		It("return true for a cloud scheduler job", func() {
			providerConfig := VAIProviderConfig{
				VaiProject:  common.RandomString(),
				VaiLocation: common.RandomString(),
			}

			legacyScheduleId := fmt.Sprintf("projects/%s/locations/%s/jobs/%s", providerConfig.VaiProject, providerConfig.VaiLocation, common.RandomString())

			Expect(isLegacySchedule(providerConfig, legacyScheduleId)).To(BeTrue())
		})

		It("return false for a vai scheduler job", func() {
			providerConfig := VAIProviderConfig{
				VaiProject:  common.RandomString(),
				VaiLocation: common.RandomString(),
			}

			scheduleId := fmt.Sprintf("projects/%s/locations/%s/schedules/%s", providerConfig.VaiProject, providerConfig.VaiLocation, common.RandomString())

			Expect(isLegacySchedule(providerConfig, scheduleId)).To(BeFalse())
		})
	})

	Describe("buildVaiSchedule", func() {
		vaiProvider := VAIProvider{}
		dummyPipelineJobRequest := &aiplatformpb.Schedule_CreatePipelineJobRequest{}

		stubBuildPipelineJob := func(ctx context.Context, providerConfig VAIProviderConfig, runScheduleDefinition RunScheduleDefinition) (*aiplatformpb.Schedule_CreatePipelineJobRequest, error) {
			return dummyPipelineJobRequest, nil
		}

		failBuildPipelineJob := func(ctx context.Context, providerConfig VAIProviderConfig, runScheduleDefinition RunScheduleDefinition) (*aiplatformpb.Schedule_CreatePipelineJobRequest, error) {
			return nil, errors.New("failed to create pipeline job request object")
		}

		It("returns schedule with fields all set as expected", func() {
			providerConfig := randomVAIProviderConfig()

			runScheduleDefinition := randomRunScheduleDefinition()
			expectedCron := "1 2 * 1 2"
			runScheduleDefinition.Schedule = expectedCron

			schedule, err := vaiProvider.buildVaiSchedule(ctx, providerConfig, runScheduleDefinition, stubBuildPipelineJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(schedule.TimeSpecification).To(Equal(&aiplatformpb.Schedule_Cron{Cron: expectedCron}))
			Expect(schedule.Request).To(Equal(dummyPipelineJobRequest))
			Expect(schedule.MaxConcurrentRunCount).To(Equal(int64(providerConfig.MaxConcurrentRunCount)))
			Expect(schedule.DisplayName).To(HavePrefix("rc-"))
		})

		It("returns error if schedule set in RunScheduleDefinition is invalid cron", func() {
			providerConfig := randomVAIProviderConfig()

			runScheduleDefinition := randomRunScheduleDefinition()
			expectedCron := "invalid cron"
			runScheduleDefinition.Schedule = expectedCron

			_, err := vaiProvider.buildVaiSchedule(ctx, providerConfig, runScheduleDefinition, stubBuildPipelineJob)
			Expect(err).To(HaveOccurred())
		})

		It("returns error if building pipeline job fails", func() {
			providerConfig := randomVAIProviderConfig()
			runScheduleDefinition := randomRunScheduleDefinition()

			_, err := vaiProvider.buildVaiSchedule(ctx, providerConfig, runScheduleDefinition, failBuildPipelineJob)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("buildPipelineJob", func() {
		newJobName := "new pipeline job name"
		mutateJob := func(ctx context.Context, config VAIProviderConfig, pipelineJob *aiplatformpb.PipelineJob) error {
			pipelineJob.Name = newJobName
			return nil
		}

		failMutateJob := func(ctx context.Context, config VAIProviderConfig, pipelineJob *aiplatformpb.PipelineJob) error {
			return errors.New("failed to mutate job")
		}

		It("should populate runtime config from run schedule parameters", func() {
			providerConfig := randomVAIProviderConfig()
			runScheduleDefinition := randomRunScheduleDefinition()
			inputParams := map[string]string{
				"name1": "value1",
				"name2": "value2",
				"name3": "value3",
			}
			runScheduleDefinition.RuntimeParameters = inputParams

			job, err := buildPipelineJob(ctx, providerConfig, runScheduleDefinition, mutateJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(job.Name).To(Equal(newJobName))
			Expect(len(job.RuntimeConfig.Parameters)).To(Equal(3))
		})

		It("returns error on failing to mutate job", func() {
			providerConfig := randomVAIProviderConfig()
			runScheduleDefinition := randomRunScheduleDefinition()

			_, err := buildPipelineJob(ctx, providerConfig, runScheduleDefinition, failMutateJob)
			Expect(err).To(HaveOccurred())
		})
	})
})
