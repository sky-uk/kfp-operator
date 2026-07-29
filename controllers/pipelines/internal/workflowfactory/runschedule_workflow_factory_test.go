//go:build unit

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/internal/config"
	"github.com/sky-uk/kfp-operator/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

var _ = Context("runConfigurationNameForRunSchedule", func() {
	Specify("returns the name of the owner if set", func() {
		runSchedule := pipelineshub.RunSchedule{}
		runSchedule.Namespace = apis.RandomString()
		runConfiguration := pipelineshub.RandomRunConfiguration(common.RandomNamespacedName())

		runSchedule.OwnerReferences = []metav1.OwnerReference{{
			Controller: pointer.Bool(true),
			APIVersion: pipelineshub.GroupVersion.String(),
			Kind:       "RunConfiguration",
			Name:       runConfiguration.Name,
		}}

		Expect(runConfigurationNameForRunSchedule(&runSchedule)).
			To(Equal(common.NamespacedName{
				Name:      runConfiguration.Name,
				Namespace: runSchedule.Namespace,
			}))
	})

	Specify("returns the empty string if owner not set", func() {
		Expect(runConfigurationNameForRunSchedule(&pipelineshub.RunSchedule{}).Empty()).To(BeTrue())
	})

	Specify("returns the empty string if the controller is not a RunConfiguration", func() {
		runSchedule := pipelineshub.RunSchedule{}

		runSchedule.OwnerReferences = append(
			runSchedule.OwnerReferences, metav1.OwnerReference{
				Controller: pointer.Bool(true),
				APIVersion: apis.RandomString(),
				Kind:       apis.RandomString(),
				Name:       apis.RandomString(),
			},
		)

		Expect(runConfigurationNameForRunSchedule(&runSchedule).Empty()).To(BeTrue())
	})
})

var _ = Describe("runScheduleDefinitionBuilder", func() {
	newRunSchedule := func() *pipelineshub.RunSchedule {
		rs := pipelineshub.RandomRunSchedule(common.RandomNamespacedName())
		rs.ObjectMeta = metav1.ObjectMeta{
			Name:      "runScheduleName",
			Namespace: "runScheduleNamespace",
		}
		return rs
	}

	Context("build", func() {
		builder := runScheduleDefinitionBuilder{config: config.ConfigSpec{}}

		When("the RunSchedule specifies an experiment name", func() {
			It("sets the run schedule namespace on the experiment name", func() {
				rs := newRunSchedule()
				rs.Spec.ExperimentName = "myExperiment"

				definition, err := builder.build(rs)
				Expect(err).NotTo(HaveOccurred())
				Expect(definition.ExperimentName).To(Equal(common.NamespacedName{
					Name:      "myExperiment",
					Namespace: "runScheduleNamespace",
				}))
			})
		})

		When("the RunSchedule does not specify an experiment name", func() {
			It("leaves the experiment name empty, without a namespace", func() {
				rs := newRunSchedule()
				rs.Spec.ExperimentName = ""

				definition, err := builder.build(rs)
				Expect(err).NotTo(HaveOccurred())
				Expect(definition.ExperimentName).To(Equal(common.NamespacedName{}))
			})
		})
	})
})
