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

type MockLabelGen struct{}

func (lg MockLabelGen) GenerateLabels(value any) (map[string]string, error) {
	switch v := value.(type) {
	case resource.RunDefinition:
		return map[string]string{
			"rd-key": "rd-value",
		}, nil
	case resource.RunScheduleDefinition:
		return map[string]string{
			"rsd-key": "rsd-value",
		}, nil
	default:
		return nil, fmt.Errorf("Unexpected value of type %T", v)
	}
}

var _ = Describe("JobBuilder", func() {
	var jb = JobBuilder{
		serviceAccount: "service-account",
		pipelineBucket: "pipeline-bucket",
		labelGen:       MockLabelGen{},
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
})
