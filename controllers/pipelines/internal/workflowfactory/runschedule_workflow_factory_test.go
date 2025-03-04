//go:build unit

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1beta1"
	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

var _ = Context("runConfigurationNameForRunSchedule", func() {
	Specify("returns the name of the owner if set", func() {
		runSchedule := pipelinesv1.RunSchedule{}
		runSchedule.Namespace = apis.RandomString()
		runConfiguration := pipelinesv1.RandomRunConfiguration(apis.RandomLowercaseString())

		runSchedule.OwnerReferences = []metav1.OwnerReference{{
			Controller: pointer.Bool(true),
			APIVersion: pipelinesv1.GroupVersion.String(),
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
		Expect(runConfigurationNameForRunSchedule(&pipelinesv1.RunSchedule{}).Empty()).To(BeTrue())
	})

	Specify("returns the empty string if the controller is not a RunConfiguration", func() {
		runSchedule := pipelinesv1.RunSchedule{}

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
