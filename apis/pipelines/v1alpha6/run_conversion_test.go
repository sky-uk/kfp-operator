//go:build unit

package v1alpha6

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
)

var _ = Context("Run Conversion", PropertyBased, func() {
	DefaultProviderNamespace = "default-provider-namespace"

	var _ = Describe("Roundtrip forward", func() {
		Specify("converts to and from the same object", func() {
			src := RandomRun(apis.RandomLowercaseString())
			intermediate := &hub.Run{}
			dst := &Run{}

			Expect(src.ConvertTo(intermediate)).To(Succeed())
			Expect(intermediate.Spec.Provider.Namespace).To(Equal(DefaultProviderNamespace))
			Expect(dst.ConvertFrom(intermediate)).To(Succeed())
			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomRun(common.RandomNamespacedName())
			intermediate := &Run{}
			dst := &hub.Run{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())
			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})
})
