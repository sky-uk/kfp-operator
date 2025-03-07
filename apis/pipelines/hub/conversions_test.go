//go:build unit

package v1beta1

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var _ = Context("Conversions", func() {
	var _ = Describe("ToTFXPipelineFramework", func() {
		Specify("returns tfx framework for empty string", func() {
			res := ToTFXPipelineFramework("")
			Expect(res.Type).To(Equal("tfx"))
			Expect(res.Parameters).To(HaveKey("components"))
		})

		Specify("returns tfx framework for some string", func() {
			tfxComponents := apis.RandomString()
			res := ToTFXPipelineFramework(tfxComponents)
			marshal, _ := json.Marshal(tfxComponents)
			Expect(res.Type).To(Equal("tfx"))
			Expect(res.Parameters["components"]).To(Equal(&apiextensionsv1.JSON{Raw: marshal}))
		})
	})

	var _ = Describe("FromPipelineFramework", func() {
		Specify("returns tfxComponents for tfx framework", func() {
			pf := PipelineFramework{
				Type:       "tfx",
				Parameters: map[string]*apiextensionsv1.JSON{"components": {Raw: []byte(`"somestring"`)}},
			}
			res, annotations, err := FromPipelineFramework(pf)
			Expect(err).NotTo(HaveOccurred())
			Expect(annotations).To(BeNil())
			Expect(res).To(Equal("somestring"))
		})

		Specify("returns error on invalid raw components value", func() {
			pf := PipelineFramework{
				Type:       "tfx",
				Parameters: map[string]*apiextensionsv1.JSON{"components": {Raw: []byte("this is not valid json")}},
			}
			res, annotations, err := FromPipelineFramework(pf)
			Expect(err).To(HaveOccurred())
			Expect(annotations).To(BeNil())
			Expect(res).To(BeEmpty())
		})

		Specify("returns error for missing components in tfx framework parameters", func() {
			pf := PipelineFramework{
				Type:       "tfx",
				Parameters: map[string]*apiextensionsv1.JSON{},
			}
			res, annotations, err := FromPipelineFramework(pf)
			Expect(err).To(HaveOccurred())
			Expect(annotations).To(BeNil())
			Expect(res).To(BeEmpty())
		})

		Specify("returns error for missing tfx framework parameters", func() {
			pf := PipelineFramework{
				Type: "tfx",
			}
			res, annotations, err := FromPipelineFramework(pf)
			Expect(err).To(HaveOccurred())
			Expect(annotations).To(BeNil())
			Expect(res).To(BeEmpty())
		})

		Specify("returns annotation for non-tfx framework", func() {
			pf := PipelineFramework{
				Type: "non-tfx",
			}
			res, annotations, err := FromPipelineFramework(pf)
			Expect(err).NotTo(HaveOccurred())
			Expect(annotations).NotTo(BeNil())
			Expect(res).To(BeEmpty())
		})
	})
})
