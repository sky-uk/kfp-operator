//go:build unit

package provider

import (
	"fmt"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/mocks"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/provider/testutil"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ = Describe("JobBuilder", func() {
	var jb = DefaultJobBuilder{
		serviceAccount: "service-account",
		pipelineBucket: "pipeline-bucket",
		labelGen:       mocks.MockLabelGen{},
	}

	Context("MkRunPipelineJob", func() {
		When("templateUri is valid", func() {
			It("should make a run pipeline job", func() {
				rd := testutil.RandomBasicRunDefinition()
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
				rd := testutil.RandomBasicRunDefinition()
				rd.PipelineName.Name = ""
				_, err := jb.MkRunPipelineJob(rd)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("MkRunSchedulePipelineJob", func() {
		When("templateUri is valid", func() {
			It("should make a run schedule pipeline job", func() {
				rsd := testutil.RandomRunScheduleDefinition()
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
				rsd := testutil.RandomRunScheduleDefinition()
				rsd.PipelineName.Name = ""
				_, err := jb.MkRunSchedulePipelineJob(rsd)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("MkSchedule", func() {
		It("should make a Schedule", func() {
			rsd := testutil.RandomRunScheduleDefinition()
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
			Expect(schedule.StartTime).To(Equal(timestamppb.New(testutil.Now.Time)))
			Expect(schedule.EndTime).To(Equal(timestamppb.New(testutil.Now.Time)))
			Expect(schedule.MaxConcurrentRunCount).To(Equal(int64(1000)))
			Expect(schedule.AllowQueueing).To(BeTrue())
		})
		When("schedule cron expression is invalid", func() {
			It("returns an error", func() {
				rsd := testutil.RandomRunScheduleDefinition()
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
				rsd := testutil.RandomRunScheduleDefinition()
				rsd.Schedule.StartTime = nil
				schedule, err := jb.MkSchedule(
					rsd,
					&aiplatformpb.PipelineJob{Name: "test"},
					"parent",
					1000,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(schedule.StartTime).To(BeNil())
				Expect(schedule.EndTime).To(Equal(timestamppb.New(testutil.Now.Time)))
			})
		})
		When("run definition schedule end time is empty", func() {
			It("should create a scheudle without end time", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				rsd.Schedule.EndTime = nil
				schedule, err := jb.MkSchedule(
					rsd,
					&aiplatformpb.PipelineJob{Name: "test"},
					"parent",
					1000,
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(schedule.StartTime).To(Equal(timestamppb.New(testutil.Now.Time)))
				Expect(schedule.EndTime).To(BeNil())
			})
		})
	})
})
