//go:build unit

package vai

import (
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/argo/providers/base"
)

func randomBasicRunDefinition() RunDefinition {
	return RunDefinition{
		Name:                 common.RandomNamespacedName(),
		PipelineName:         common.RandomNamespacedName(),
		PipelineVersion:      common.RandomString(),
		RunConfigurationName: common.RandomNamespacedName(),
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
		PipelineBucket:        common.RandomString(),
		VaiProject:            common.RandomString(),
		VaiLocation:           common.RandomString(),
		MaxConcurrentRunCount: common.RandomInt64(),
	}
}

var _ = Context("VAI Provider", func() {

	DescribeTable("extractBucketAndObjectFromGCSPath", func(in string, expectedBucket string, expectedPath string, expectedErr error) {
		actualBucket, actualPath, actualError := extractBucketAndObjectFromGCSPath(in)
		Expect(actualBucket).To(Equal(expectedBucket))
		Expect(actualPath).To(Equal(expectedPath))
		if expectedErr != nil {
			Expect(actualError).To(Equal(expectedErr))
		} else {
			Expect(actualError).ToNot(HaveOccurred())
		}
	}, Entry("", "", "", "", errors.New("invalid gs URI []")),
		Entry("", "not valid", "", "", errors.New("invalid gs URI [not valid]")),
		Entry("", "gs://", "", "", errors.New("invalid gs URI [gs://]")),
		Entry("", "gs://bucket", "", "", errors.New("invalid gs URI [gs://bucket]")),
		Entry("", "gs://bucket/path", "bucket", "path", nil),
		Entry("", "gs://bucket/path/still-path/more", "bucket", "path/still-path/more", nil),
	)

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
			input := randomBasicRunDefinition()
			input.PipelineVersion = "0.4.0"

			runLabels := runLabelsFromRunDefinition(input)

			Expect(runLabels[labels.PipelineVersion]).To(Equal("0-4-0"))
		})
	})

	Describe("runLabelsFromRun", func() {
		It("generates run labels with RunName", func() {
			input := randomBasicRunDefinition()
			input.RunConfigurationName = common.NamespacedName{}

			runLabels := runLabelsFromRunDefinition(input)

			Expect(runLabels[labels.RunName]).To(Equal(input.Name.Name))
			Expect(runLabels[labels.RunNamespace]).To(Equal(input.Name.Namespace))
			Expect(runLabels).NotTo(HaveKey(labels.RunConfigurationName))
			Expect(runLabels).NotTo(HaveKey(labels.RunConfigurationNamespace))
		})

		It("generates run labels with RunConfigurationName and RunName", func() {
			input := randomBasicRunDefinition()

			runLabels := runLabelsFromRunDefinition(input)

			Expect(runLabels[labels.RunConfigurationName]).To(Equal(input.RunConfigurationName.Name))
			Expect(runLabels[labels.RunConfigurationNamespace]).To(Equal(input.RunConfigurationName.Namespace))
			Expect(runLabels[labels.RunName]).To(Equal(input.Name.Name))
			Expect(runLabels[labels.RunNamespace]).To(Equal(input.Name.Namespace))
		})

		It("replaces fullstops with dashes in pipelineVersion", func() {
			input := randomBasicRunDefinition()
			input.PipelineVersion = "0.4.0"

			runLabels := runLabelsFromRunDefinition(input)

			Expect(runLabels[labels.PipelineVersion]).To(Equal("0-4-0"))
		})
	})

	vaiProvider := VAIProvider{}

	Describe("buildVaiSchedule", func() {
		emptyPipelineJob := aiplatformpb.PipelineJob{}

		It("returns schedule with fields all set as expected", func() {
			providerConfig := randomVAIProviderConfig()

			runScheduleDefinition := randomRunScheduleDefinition()
			expectedCron := "1 2 * 1 2"
			runScheduleDefinition.Schedule = expectedCron

			schedule, err := vaiProvider.buildVaiScheduleFromPipelineJob(providerConfig, runScheduleDefinition, &emptyPipelineJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(schedule.TimeSpecification).To(Equal(&aiplatformpb.Schedule_Cron{Cron: expectedCron}))
			Expect(schedule.MaxConcurrentRunCount).To(Equal(providerConfig.getMaxConcurrentRunCountOrDefault()))
			Expect(schedule.DisplayName).To(HavePrefix("rc-"))
		})

		It("returns error if schedule set in RunScheduleDefinition is invalid cron", func() {
			providerConfig := randomVAIProviderConfig()

			runScheduleDefinition := randomRunScheduleDefinition()
			expectedCron := "invalid cron"
			runScheduleDefinition.Schedule = expectedCron

			_, err := vaiProvider.buildVaiScheduleFromPipelineJob(providerConfig, runScheduleDefinition, &emptyPipelineJob)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("buildPipelineJob", func() {

		It("should populate runtime config from run schedule parameters", func() {
			providerConfig := randomVAIProviderConfig()
			runScheduleDefinition := randomRunScheduleDefinition()
			inputParams := map[string]string{
				"name1": "value1",
				"name2": "value2",
				"name3": "value3",
			}
			runScheduleDefinition.RuntimeParameters = inputParams

			job, err := vaiProvider.buildPipelineJob(providerConfig, runScheduleDefinition)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(job.RuntimeConfig.Parameters)).To(Equal(3))
			Expect(job.RuntimeConfig.Parameters["name1"]).To(Equal(&aiplatformpb.Value{
				Value: &aiplatformpb.Value_StringValue{
					StringValue: "value1",
				},
			}))
			Expect(job.RuntimeConfig.Parameters["name2"]).To(Equal(&aiplatformpb.Value{
				Value: &aiplatformpb.Value_StringValue{
					StringValue: "value2",
				},
			}))
			Expect(job.RuntimeConfig.Parameters["name3"]).To(Equal(&aiplatformpb.Value{
				Value: &aiplatformpb.Value_StringValue{
					StringValue: "value3",
				},
			}))
		})
	})

	Describe("ExtractFromMap", func() {
		It("should extract a value from a map for any given type if it exists", func() {
			testMap := map[string]any{
				"stringField": "someString",
				"intField":    1,
				"mapField": map[string]int{
					"nestedField": 2,
				},
			}

			stringField, stringFieldErr := ExtractFromMap[string](testMap, "stringField")
			intField, intFieldErr := ExtractFromMap[int](testMap, "intField")
			mapField, mapFieldErr := ExtractFromMap[map[string]int](testMap, "mapField")
			_, missingFieldErr := ExtractFromMap[string](testMap, "missingField")
			_, wrongTypeError := ExtractFromMap[int](testMap, "mapField")

			Expect(stringField).To(Equal("someString"))
			Expect(stringFieldErr).ToNot(HaveOccurred())
			Expect(intField).To(Equal(1))
			Expect(intFieldErr).ToNot(HaveOccurred())
			Expect(mapField).To(Equal(map[string]int{
				"nestedField": 2,
			}))
			Expect(mapFieldErr).ToNot(HaveOccurred())
			Expect(missingFieldErr).To(HaveOccurred())
			Expect(wrongTypeError).To(HaveOccurred())
		})
	})

	Describe("retrieveRunIdFromSpec", func() {
		It("should extract a runId from a given pipelineSpec", func() {
			pipelineName := "pipelineName"
			pipelineSpec := map[string]any{
				"pipelineInfo": map[string]any{
					"name": pipelineName,
				},
			}
			result, err := retrieveRunIdFromSpec(pipelineSpec)

			Expect(result).To(Equal(pipelineName))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should error if pipeline spec is missing a required field", func() {
			pipelineSpec := map[string]any{}
			_, err := retrieveRunIdFromSpec(pipelineSpec)

			Expect(err).To(HaveOccurred())
		})
	})
})
