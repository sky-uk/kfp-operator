//go:build unit

package util

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/common"
)

func TestGcsUtilUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GCS Util Unit Suite")
}

var _ = Describe("Util", Ordered, func() {
	DescribeTable(
		"PipelineStorageObject",
		func(
			pipelineName common.NamespacedName,
			pipelineVersion string,
			expectedStorageObject string,
		) {
			storageObject, err := PipelineStorageObject(pipelineName, pipelineVersion)
			if expectedStorageObject == "" {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(storageObject).To(Equal(expectedStorageObject))
			}
		},
		Entry(
			"namespace and name non-empty",
			common.NamespacedName{
				Name:      "myName",
				Namespace: "myNamespace",
			},
			"version", "myNamespace/myName/version",
		),
		Entry(
			"empty namespace",
			common.NamespacedName{
				Name:      "myName",
				Namespace: "",
			},
			"version", "myName/version",
		),
		Entry(
			"empty name",
			common.NamespacedName{
				Name:      "",
				Namespace: "myNamespace",
			},
			"version",
			"",
		),
	)

	DescribeTable(
		"PipelineUri",
		func(
			bucket string,
			pipelineName common.NamespacedName,
			pipelineVersion string,
			expectedStorageObject string,
		) {
			storageObject, err := PipelineUri(pipelineName, pipelineVersion, bucket)
			if expectedStorageObject == "" {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(storageObject).To(Equal(expectedStorageObject))
			}
		},
		Entry(
			"namespace, name and bucket non-empty",
			"bucket",
			common.NamespacedName{
				Name:      "myName",
				Namespace: "myNamespace",
			},
			"version",
			"gs://bucket/myNamespace/myName/version",
		),
		Entry(
			"empty bucket",
			"",
			common.NamespacedName{
				Name:      "myName",
				Namespace: "myNamespace",
			},
			"version",
			"gs:///myNamespace/myName/version",
		),
		Entry(
			"empty namespace",
			"bucket",
			common.NamespacedName{
				Name:      "myName",
				Namespace: "",
			},
			"version",
			"gs://bucket/myName/version",
		),
		Entry(
			"empty name",
			"bucket",
			common.NamespacedName{
				Name:      "",
				Namespace: "myNamespace",
			},
			"version",
			"",
		),
	)
})
