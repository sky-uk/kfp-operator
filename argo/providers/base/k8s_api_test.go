//go:build unit

package base

import (
	"context"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Context("K8sApi", func() {
	ctx := logr.NewContext(context.Background(), logr.Discard())

	Describe("artifactsForUnstructured", func() {
		unstructuredArtifacts := make([]interface{}, 2)
		name1 := "somename"
		name2 := "someothername"
		path1Str := "Pusher:pushed_model:0[pushed == 1]"
		path2Str := "Pusher:pushed_model:1[pushed == 2]"

		unstructuredArtifacts[0] = map[string]interface{}{
			"name": name1,
			"path": path1Str,
		}
		unstructuredArtifacts[1] = map[string]interface{}{
			"name": name2,
			"path": path2Str,
		}

		expectedArtifacts := make([]v1alpha5.OutputArtifact, 2)
		path1, err := v1alpha5.ArtifactPathFromString(path1Str)
		Expect(err).NotTo(HaveOccurred())
		path2, err := v1alpha5.ArtifactPathFromString(path2Str)
		Expect(err).NotTo(HaveOccurred())
		expectedArtifacts[0] = v1alpha5.OutputArtifact{
			Name: name1,
			Path: path1,
		}
		expectedArtifacts[1] = v1alpha5.OutputArtifact{
			Name: name2,
			Path: path2,
		}

		When("a run has artifacts", func() {
			It("returns the artifacts", func() {
				obj := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{
							"artifacts": unstructuredArtifacts,
						},
					},
				}

				artifacts, err := artifactsForUnstructured(ctx, obj, RunGVR)

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

				artifacts, err := artifactsForUnstructured(ctx, obj, RunGVR)
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

				_, err := artifactsForUnstructured(ctx, obj, RunGVR)
				Expect(err).To(HaveOccurred())
			})
		})

		When("a run configuration has artifacts", func() {
			It("returns the artifacts", func() {
				obj := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{
							"run": map[string]interface{}{
								"artifacts": unstructuredArtifacts,
							},
						},
					},
				}

				artifacts, err := artifactsForUnstructured(ctx, obj, RunConfigurationGVR)
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

				artifacts, err := artifactsForUnstructured(ctx, obj, RunConfigurationGVR)
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

				_, err := artifactsForUnstructured(ctx, obj, RunConfigurationGVR)
				Expect(err).To(HaveOccurred())
			})
		})

		When("an unknown resource is provided", func() {
			It("errors", func() {
				obj := &unstructured.Unstructured{
					Object: map[string]interface{}{},
				}

				_, err := artifactsForUnstructured(ctx, obj, schema.GroupVersionResource{Group: apis.RandomString(), Version: apis.RandomString(), Resource: apis.RandomString()})
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
