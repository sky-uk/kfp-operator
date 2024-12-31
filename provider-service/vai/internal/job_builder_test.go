//go:build unit

package internal

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	var jb = JobBuilder{
		serviceAccount: "service-account",
		pipelineBucket: "pipeline-bucket",
	}

	Context("runLabelsFromRunDefinition", func() {
		When("RunConfigurationName and RunName is present", func() {
			It("generates run labels with RunConfigurationName and RunName", func() {
				rd := randomBasicRunDefinition()
				rl := jb.runLabelsFromRunDefinition(rd)

				Expect(rl[labels.PipelineName]).To(Equal(rd.PipelineName.Name))
				Expect(rl[labels.PipelineNamespace]).To(Equal(rd.PipelineName.Namespace))
				Expect(rl[labels.PipelineVersion]).To(Equal(rd.PipelineVersion))
				Expect(rl[labels.RunConfigurationName]).To(Equal(rd.RunConfigurationName.Name))
				Expect(rl[labels.RunConfigurationNamespace]).To(Equal(rd.RunConfigurationName.Namespace))
				Expect(rl[labels.RunName]).To(Equal(rd.Name.Name))
				Expect(rl[labels.RunNamespace]).To(Equal(rd.Name.Namespace))
			})
		})
		When("RunConfigurationName is empty", func() {
			It("generates run labels with RunName", func() {
				rd := randomBasicRunDefinition()
				rd.RunConfigurationName = common.NamespacedName{}
				rl := jb.runLabelsFromRunDefinition(rd)

				Expect(rl[labels.PipelineName]).To(Equal(rd.PipelineName.Name))
				Expect(rl[labels.PipelineNamespace]).To(Equal(rd.PipelineName.Namespace))
				Expect(rl[labels.PipelineVersion]).To(Equal(rd.PipelineVersion))
				Expect(rl[labels.RunName]).To(Equal(rd.Name.Name))
				Expect(rl[labels.RunNamespace]).To(Equal(rd.Name.Namespace))
				Expect(rl).NotTo(HaveKey(labels.RunConfigurationName))
				Expect(rl).NotTo(HaveKey(labels.RunConfigurationNamespace))
			})
		})
		When("RunName is empty", func() {
			It("generates run labels with RunName", func() {
				rd := randomBasicRunDefinition()
				rd.Name = common.NamespacedName{}
				rl := jb.runLabelsFromRunDefinition(rd)

				Expect(rl[labels.PipelineName]).To(Equal(rd.PipelineName.Name))
				Expect(rl[labels.PipelineNamespace]).To(Equal(rd.PipelineName.Namespace))
				Expect(rl[labels.PipelineVersion]).To(Equal(rd.PipelineVersion))
				Expect(rl[labels.RunConfigurationName]).To(Equal(rd.RunConfigurationName.Name))
				Expect(rl[labels.RunConfigurationNamespace]).To(Equal(rd.RunConfigurationName.Namespace))
				Expect(rl).NotTo(HaveKey(labels.RunName))
				Expect(rl).NotTo(HaveKey(labels.RunNamespace))
			})
		})
		It("replaces fullstops with dashes in pipelineVersion", func() {
			rd := randomBasicRunDefinition()
			rd.PipelineVersion = "0.4.0"
			rl := jb.runLabelsFromRunDefinition(rd)

			Expect(rl[labels.PipelineVersion]).To(Equal("0-4-0"))
		})
	})

	Context("runLabelsFromSchedule", func() {
		When("RunConfigurationName is present", func() {
			It("generates run labels with RunConfiguration name and namespace", func() {
				rsd := randomRunScheduleDefinition()
				rl := jb.runLabelsFromSchedule(rsd)

				Expect(rl[labels.PipelineName]).To(Equal(rsd.PipelineName.Name))
				Expect(rl[labels.PipelineNamespace]).To(Equal(rsd.PipelineName.Namespace))
				Expect(rl[labels.PipelineVersion]).To(Equal(rsd.PipelineVersion))
				Expect(rl[labels.RunConfigurationName]).To(Equal(rsd.RunConfigurationName.Name))
				Expect(rl[labels.RunConfigurationNamespace]).To(Equal(rsd.RunConfigurationName.Namespace))
			})
		})
		When("RunConfigurationName is empty", func() {
			It("generates run labels without RunConfiguration name and namespace", func() {
				rsd := randomRunScheduleDefinition()
				rsd.RunConfigurationName = common.NamespacedName{}
				rl := jb.runLabelsFromSchedule(rsd)

				Expect(rl[labels.PipelineName]).To(Equal(rsd.PipelineName.Name))
				Expect(rl[labels.PipelineNamespace]).To(Equal(rsd.PipelineName.Namespace))
				Expect(rl[labels.PipelineVersion]).To(Equal(rsd.PipelineVersion))
				Expect(rl).NotTo(HaveKey(labels.RunConfigurationName))
				Expect(rl).NotTo(HaveKey(labels.RunConfigurationNamespace))
			})
		})

		It("replaces fullstops with dashes in pipelineVersion", func() {
			rd := randomBasicRunDefinition()
			rd.PipelineVersion = "0.4.0"
			rl := jb.runLabelsFromRunDefinition(rd)

			Expect(rl[labels.PipelineVersion]).To(Equal("0-4-0"))
		})
	})

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
				for k, v := range job.RuntimeConfig.Parameters {
					Expect(v.GetStringValue).To(Equal(rd.RuntimeParameters[k]))
				}
				// TODO: assert labels
				Expect(job.ServiceAccount).To(Equal(jb.serviceAccount))
				Expect(job.TemplateUri).To(Equal(expectedTemplateUri))
			})
		})
		When("templateUri is invalid", func() {
			It("should return erro", func() {
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
				Expect(job.ServiceAccount).To(Equal(jb.serviceAccount))
				Expect(job.TemplateUri).To(Equal(expectedTemplateUri))
			})
		})
	})
})
