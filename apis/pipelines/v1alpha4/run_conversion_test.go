//go:build unit
// +build unit

package v1alpha4

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
)

var _ = Context("Run Conversion", PropertyBased, func() {
	var _ = Describe("Roundtrip", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomRun()
			hub.WithValueFrom(&src.Spec)
			intermediate := &Run{}
			dst := &hub.Run{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(dst.Spec.RuntimeParameters).To(ConsistOf(src.Spec.RuntimeParameters))
			dst.Spec.RuntimeParameters = nil
			src.Spec.RuntimeParameters = nil
			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})
})
