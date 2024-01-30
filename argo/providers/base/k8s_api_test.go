//go:build unit

package base

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Context("K8sApi", func() {
	Describe("artifactsForUnstructured", func() {
		When("a run has artifacts", func() {
			It("returns the artifacts", func() {
				expectedArtifacts := make([]map[string]string, 2)
				expectedArtifacts[0] = map[string]string{
					"name": "somename",
					"path": "Pusher:pushed_model:0[pushed == 1]",
				}
				expectedArtifacts[1] = map[string]string{
					"name": "someother",
					"path": "Pusher:pushed_model:1[pushed == 2]",
				}

				obj := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{
							"artifacts": expectedArtifacts,
						},
					},
				}

				artifactsField, hasArtifacts, err := unstructured.NestedFieldNoCopy(obj.Object, "spec", "artifacts")
				Expect(err).NotTo(HaveOccurred())
				Expect(hasArtifacts).To(BeTrue())
				println(fmt.Sprintf("artifactsField:%+v", artifactsField))

				flattenedArtifacts := artifactsField.([]map[string]string)
				println(fmt.Sprintf("FLATTENED:%+v", flattenedArtifacts))
				//println(fmt.Sprintf("OK:%+v", ok))
				Expect(true).To(BeFalse())

				artifacts, err := artifactsForUnstructured(obj, RunGVR)
				Expect(err).NotTo(HaveOccurred())
				Expect(artifacts).To(Equal(expectedArtifacts))
			})
		})

		When("a run has no artifacts", func() {
			It("returns no artifacts", func() {
				obj := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{},
					},
				}

				artifacts, err := artifactsForUnstructured(obj, RunGVR)
				Expect(err).NotTo(HaveOccurred())
				Expect(artifacts).To(BeEmpty())
			})
		})

		When("a run is malformed", func() {
			It("errors", func() {
				obj := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{
							"artifacts": apis.RandomString(),
						},
					},
				}

				_, err := artifactsForUnstructured(obj, RunGVR)
				Expect(err).To(HaveOccurred())
			})
		})

		When("a run configuration has artifacts", func() {
			It("returns the artifacts", func() {
				expectedArtifacts := apis.RandomList(pipelinesv1.RandomOutputArtifact)

				obj := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{
							"run": map[string]interface{}{
								"artifacts": expectedArtifacts,
							},
						},
					},
				}

				artifacts, err := artifactsForUnstructured(obj, RunConfigurationGVR)
				Expect(err).NotTo(HaveOccurred())
				Expect(artifacts).To(Equal(expectedArtifacts))
			})
		})

		When("a run configuration has no artifacts", func() {
			It("returns no artifacts", func() {
				obj := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"run": map[string]interface{}{},
					},
				}

				artifacts, err := artifactsForUnstructured(obj, RunConfigurationGVR)
				Expect(err).NotTo(HaveOccurred())
				Expect(artifacts).To(BeEmpty())
			})
		})

		When("a run configuration is malformed", func() {
			It("errors", func() {
				obj := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{
							"run": map[string]interface{}{
								"artifacts": apis.RandomString(),
							},
						},
					},
				}

				_, err := artifactsForUnstructured(obj, RunConfigurationGVR)
				Expect(err).To(HaveOccurred())
			})
		})

		When("an unknown resource is provided", func() {
			It("errors", func() {
				obj := &unstructured.Unstructured{
					Object: map[string]interface{}{},
				}

				_, err := artifactsForUnstructured(obj, schema.GroupVersionResource{Group: apis.RandomString(), Version: apis.RandomString(), Resource: apis.RandomString()})
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
