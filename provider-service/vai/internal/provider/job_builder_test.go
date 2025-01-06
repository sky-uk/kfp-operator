//go:build unit

package provider

import (
	"fmt"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/mocks"
	"google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO extract to different file as it's shared
func randomBasicRunDefinition() resource.RunDefinition {
	return resource.RunDefinition{
		Name:                 common.RandomNamespacedName(),
		PipelineName:         common.RandomNamespacedName(),
		PipelineVersion:      common.RandomString(),
		RunConfigurationName: common.RandomNamespacedName(),
	}
}

var now = metav1.Now()

func randomRunScheduleDefinition() resource.RunScheduleDefinition {
	return resource.RunScheduleDefinition{
		Name:                 common.RandomNamespacedName(),
		Version:              common.RandomString(),
		PipelineName:         common.RandomNamespacedName(),
		PipelineVersion:      common.RandomString(),
		RunConfigurationName: common.RandomNamespacedName(),
		ExperimentName:       common.RandomNamespacedName(),
		Schedule: pipelinesv1.Schedule{
			CronExpression: "1 1 0 0 0",
			StartTime:      &now,
			EndTime:        &now,
		},
	}
}

var _ = Describe("JobBuilder", func() {
	var jb = DefaultJobBuilder{
		serviceAccount: "service-account",
		pipelineBucket: "pipeline-bucket",
		labelGen:       mocks.MockLabelGen{},
	}

	Context("MkRunPipelineJob", func() {
		When("templateUri is valid", func() {
			It("should make a run pipeline job", func() {
				rd := randomBasicRunDefinition()
				job, err := jb.MkRunPipelineJob(rd)
				expectedTemplateUri := fmt.Sprintf(
					"gs://%s/%s/%s/%s",
					jb.pipelineBucket,
					rd.PipelineName.Namespace,
					rd.PipelineName.Name,
					rd.PipelineVersion,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(job.Labels).To(Equal(map[string]string{"rd-key": "rd-value"}))
				for k, v := range job.RuntimeConfig.Parameters {
					Expect(v.GetStringValue).To(Equal(rd.RuntimeParameters[k]))
				}
				Expect(job.ServiceAccount).To(Equal(jb.serviceAccount))
				Expect(job.TemplateUri).To(Equal(expectedTemplateUri))
			})
		})
		When("templateUri is invalid", func() {
			It("should return error", func() {
				rd := randomBasicRunDefinition()
				rd.PipelineName.Name = ""
				_, err := jb.MkRunPipelineJob(rd)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("MkRunSchedulePipelineJob", func() {
		When("templateUri is valid", func() {
			It("should make a run schedule pipeline job", func() {
				rsd := randomRunScheduleDefinition()
				job, err := jb.MkRunSchedulePipelineJob(rsd)
				expectedTemplateUri := fmt.Sprintf(
					"gs://%s/%s/%s/%s",
					jb.pipelineBucket,
					rsd.PipelineName.Namespace,
					rsd.PipelineName.Name,
					rsd.PipelineVersion,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(job.Labels).To(Equal(map[string]string{"rsd-key": "rsd-value"}))
				for k, v := range job.RuntimeConfig.Parameters {
					Expect(v.GetStringValue).To(Equal(rsd.RuntimeParameters[k]))
				}
				Expect(job.ServiceAccount).To(Equal(jb.serviceAccount))
				Expect(job.TemplateUri).To(Equal(expectedTemplateUri))
			})
		})
		When("templateUri is invalid", func() {
			It("should return error", func() {
				rsd := randomRunScheduleDefinition()
				rsd.PipelineName.Name = ""
				_, err := jb.MkRunSchedulePipelineJob(rsd)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("MkSchedule", func() {
		It("should make a Schedule", func() {
			rsd := randomRunScheduleDefinition()
			expectedCron := aiplatformpb.Schedule_Cron{Cron: "1 2 * 1 2"}
			rsd.Schedule.CronExpression = expectedCron.Cron

			expectedPipelineJob := aiplatformpb.PipelineJob{Name: "test"}
			schedule, err := jb.MkSchedule(
				rsd,
				&expectedPipelineJob,
				"parent",
				1000,
			)

			expectedPipelineJobReq := aiplatformpb.Schedule_CreatePipelineJobRequest{
				CreatePipelineJobRequest: &aiplatformpb.CreatePipelineJobRequest{
					Parent:      "parent",
					PipelineJob: &expectedPipelineJob,
				},
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(schedule.TimeSpecification).To(Equal(&expectedCron))
			Expect(schedule.Request).To(Equal(&expectedPipelineJobReq))
			Expect(schedule.DisplayName).To(Equal(fmt.Sprintf("rc-%s-%s", rsd.Name.Namespace, rsd.Name.Name)))
			Expect(schedule.StartTime).To(Equal(timestamppb.New(now.Time)))
			Expect(schedule.EndTime).To(Equal(timestamppb.New(now.Time)))
			Expect(schedule.MaxConcurrentRunCount).To(Equal(int64(1000)))
			Expect(schedule.AllowQueueing).To(BeTrue())
		})
		When("schedule cron expression is invalid", func() {
			It("returns an error", func() {
				rsd := randomRunScheduleDefinition()
				rsd.Schedule.CronExpression = "invalid cron"

				_, err := jb.MkSchedule(
					rsd,
					&aiplatformpb.PipelineJob{Name: "test"},
					"parent",
					1000,
				)
				Expect(err).To(HaveOccurred())
			})
		})
		When("run definition schedule start time is empty", func() {
			It("should create a scheudle without start time", func() {
				rsd := randomRunScheduleDefinition()
				rsd.Schedule.StartTime = nil
				schedule, err := jb.MkSchedule(
					rsd,
					&aiplatformpb.PipelineJob{Name: "test"},
					"parent",
					1000,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(schedule.StartTime).To(BeNil())
				Expect(schedule.EndTime).To(Equal(timestamppb.New(now.Time)))
			})
		})
		When("run definition schedule end time is empty", func() {
			It("should create a scheudle without end time", func() {
				rsd := randomRunScheduleDefinition()
				rsd.Schedule.EndTime = nil
				schedule, err := jb.MkSchedule(
					rsd,
					&aiplatformpb.PipelineJob{Name: "test"},
					"parent",
					1000,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(schedule.StartTime).To(Equal(timestamppb.New(now.Time)))
				Expect(schedule.EndTime).To(BeNil())
			})
		})
	})
})
