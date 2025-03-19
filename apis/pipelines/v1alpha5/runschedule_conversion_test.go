//go:build unit

package v1alpha5

import (
	"fmt"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
)

var _ = Context("RunSchedule Conversion", PropertyBased, func() {
	DefaultProvider = "default-provider"
	DefaultProviderNamespace = "default-provider-namespace"

	var _ = Describe("Roundtrip forward", func() {
		Specify("converts to and from the same object using default provider", func() {
			src := RandomRunSchedule()
			intermediate := &hub.RunSchedule{}
			dst := &RunSchedule{}

			Expect(src.ConvertTo(intermediate)).To(Succeed())

			fmt.Println("intermediate.Spec.Provider.Name: ", intermediate.Spec.Provider.Name)
			fmt.Println("intermediate.Spec.Provider.Namespace: ", intermediate.Spec.Provider.Namespace)
			Expect(intermediate.Spec.Provider.Name).To(Equal(DefaultProvider))
			Expect(intermediate.Spec.Provider.Namespace).To(Equal(DefaultProviderNamespace))
			Expect(dst.ConvertFrom(intermediate)).To(Succeed())
			delete(
				dst.GetAnnotations(),
				RunScheduleConversionRemainder{}.ConversionAnnotation(),
			)
			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomRunSchedule(common.RandomNamespacedName())
			intermediate := &RunSchedule{}
			dst := &hub.RunSchedule{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())
			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})
})
