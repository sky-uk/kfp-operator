//go:build unit

package vai

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
)

var _ = Context("VAI Config", func() {

	config := VAIProviderConfig{}

	DescribeTable("getMaxConcurrentRunCountOrDefault", func(setValue int64, expectDefault bool) {
		config.MaxConcurrentRunCount = setValue
		expectedRes := setValue
		if expectDefault {
			expectedRes = 10
		}

		Expect(config.getMaxConcurrentRunCountOrDefault()).To(Equal(expectedRes))

	},
		Entry("", int64(0), true),
		Entry("", -common.RandomInt64(), true),
		Entry("", common.RandomInt64()+1, false),
	)

	DescribeTable("pipelineStorageObject", func(pipelineName common.NamespacedName, pipelineVersion string, expectedStorageObject string) {
		storageObject, err := config.pipelineStorageObject(pipelineName, pipelineVersion)
		if expectedStorageObject == "" {
			Expect(err).To(HaveOccurred())
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(storageObject).To(Equal(expectedStorageObject))
		}
	},
		Entry("", common.NamespacedName{
			Name:      "myName",
			Namespace: "myNamespace",
		}, "version", "myNamespace/myName/version"),
		Entry("", common.NamespacedName{
			Name:      "myName",
			Namespace: "",
		}, "version", "myName/version"),
		Entry("", common.NamespacedName{
			Name:      "",
			Namespace: "myNamespace",
		}, "version", ""),
	)

	DescribeTable("pipelineUri", func(bucket string, pipelineName common.NamespacedName, pipelineVersion string, expectedStorageObject string) {
		config.PipelineBucket = bucket

		storageObject, err := config.pipelineUri(pipelineName, pipelineVersion)
		if expectedStorageObject == "" {
			Expect(err).To(HaveOccurred())
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(storageObject).To(Equal(expectedStorageObject))
		}
	},
		Entry("", "bucket", common.NamespacedName{
			Name:      "myName",
			Namespace: "myNamespace",
		}, "version", "gs://bucket/myNamespace/myName/version"),
		Entry("", "", common.NamespacedName{
			Name:      "myName",
			Namespace: "myNamespace",
		}, "version", "gs:///myNamespace/myName/version"),
		Entry("", "bucket", common.NamespacedName{
			Name:      "",
			Namespace: "myNamespace",
		}, "version", ""),
	)
})
