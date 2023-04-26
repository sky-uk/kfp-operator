//go:build unit
// +build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

var _ = Context("runConfigurationNameForRunSchedule", func() {
	Specify("returns the name of the owner if set", func() {
		runSchedule := pipelinesv1.RunSchedule{}
		runConfiguration := pipelinesv1.RandomRunConfiguration()

		runSchedule.OwnerReferences = []metav1.OwnerReference{{
			Controller: pointer.Bool(true),
			APIVersion: pipelinesv1.GroupVersion.String(),
			Kind:       "RunConfiguration",
			Name:       runConfiguration.Name,
		}}

		Expect(runConfigurationNameForRunSchedule(&runSchedule)).To(Equal(runConfiguration.Name))
	})

	Specify("returns the empty string if owner not set", func() {
		Expect(runConfigurationNameForRunSchedule(&pipelinesv1.RunSchedule{})).To(BeEmpty())
	})

	Specify("returns the empty string if the controller is not a RunConfiguration", func() {
		runSchedule := pipelinesv1.RunSchedule{}

		runSchedule.OwnerReferences = append(runSchedule.OwnerReferences, metav1.OwnerReference{
			Controller: pointer.Bool(true),
			APIVersion: apis.RandomString(),
			Kind:       apis.RandomString(),
			Name:       apis.RandomString(),
		})

		Expect(runConfigurationNameForRunSchedule(&runSchedule)).To(BeEmpty())
	})
})
