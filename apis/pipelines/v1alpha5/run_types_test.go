//go:build unit
// +build unit

package v1alpha5

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
)

var _ = Context("Run", func() {
	var _ = Describe("ComputeHash", func() {
		Specify("Pipeline should change the hash", func() {
			run := Run{}
			hash1 := run.ComputeHash()

			run.Spec.Pipeline = PipelineIdentifier{Name: "notempty"}
			hash2 := run.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("ExperimentName should change the hash", func() {
			run := Run{}
			hash1 := run.ComputeHash()

			run.Spec.ExperimentName = "notempty"
			hash2 := run.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("ObservedPipelineVersion should change the hash", func() {
			run := Run{}
			hash1 := run.ComputeHash()

			run.Status.ObservedPipelineVersion = "notempty"
			hash2 := run.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("All RuntimeParameters keys should change the hash", func() {
			run := Run{}
			hash1 := run.ComputeHash()

			run.Spec.RuntimeParameters = []RuntimeParameter{
				{Name: "a", Value: ""},
			}
			hash2 := run.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))

			run.Spec.RuntimeParameters = []RuntimeParameter{
				{Name: "b", Value: "notempty"},
			}
			hash3 := run.ComputeHash()

			Expect(hash2).NotTo(Equal(hash3))
		})

		Specify("The original object should not change", PropertyBased, func() {
			run := RandomRun()
			expected := run.DeepCopy()
			run.ComputeHash()

			Expect(run).To(Equal(expected))
		})
	})

	var _ = Describe("ComputeVersion", func() {
		Specify("Should have the spec hash only", func() {
			Expect(Run{}.ComputeVersion()).To(MatchRegexp("^[a-z0-9]{6}$"))
		})
	})

	Context("cmpRuntimeParameters", func() {
		unchanged := apis.RandomString()

		DescribeTable("cmpRuntimeParameters",
			func(rp1, rp2 RuntimeParameter, expected bool) {
				Expect(cmpRuntimeParameters(rp1, rp2)).To(Equal(expected))
			},
			Entry("first name less than",
				RuntimeParameter{Name: "A", Value: unchanged},
				RuntimeParameter{Name: "B", Value: unchanged},
				true),
			Entry("first name greater than",
				RuntimeParameter{Name: "B", Value: unchanged},
				RuntimeParameter{Name: "A", Value: unchanged},
				false),
			Entry("first value less than",
				RuntimeParameter{Name: unchanged, Value: "A"},
				RuntimeParameter{Name: unchanged, Value: "B"},
				true),
			Entry("first value greater than",
				RuntimeParameter{Name: unchanged, Value: "B"},
				RuntimeParameter{Name: unchanged, Value: "A"},
				false),
			Entry("same value",
				RuntimeParameter{Name: unchanged, Value: unchanged},
				RuntimeParameter{Name: unchanged, Value: unchanged},
				false),
			Entry("first runconfiguration name less than",
				RuntimeParameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: "A", OutputArtifact: unchanged}}},
				RuntimeParameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: "B", OutputArtifact: unchanged}}},
				true),
			Entry("first runconfiguration name greater than",
				RuntimeParameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: "B", OutputArtifact: unchanged}}},
				RuntimeParameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: "A", OutputArtifact: unchanged}}},
				false),
			Entry("first runconfiguration outputArtifact less than",
				RuntimeParameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: unchanged, OutputArtifact: "A"}}},
				RuntimeParameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: unchanged, OutputArtifact: "B"}}},
				true),
			Entry("first runconfiguration outputArtifact greater than",
				RuntimeParameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: unchanged, OutputArtifact: "B"}}},
				RuntimeParameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: unchanged, OutputArtifact: "A"}}},
				false),
			Entry("same valueFrom",
				RuntimeParameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: unchanged, OutputArtifact: unchanged}}},
				RuntimeParameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: unchanged, OutputArtifact: unchanged}}},
				false),
		)
	})

	Context("writeRuntimeParameter", func() {
		Specify("name should change the hash", func() {
			oh1 := pipelines.NewObjectHasher()
			rp := RuntimeParameter{}
			writeRuntimeParameter(oh1, rp)

			oh2 := pipelines.NewObjectHasher()
			rp.Name = apis.RandomString()
			writeRuntimeParameter(oh2, rp)

			Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
		})

		Specify("value should change the hash", func() {
			oh1 := pipelines.NewObjectHasher()
			rp := RuntimeParameter{}
			writeRuntimeParameter(oh1, rp)

			oh2 := pipelines.NewObjectHasher()
			rp.Value = apis.RandomString()
			writeRuntimeParameter(oh2, rp)

			Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
		})

		Specify("name in valueFrom should change the hash", func() {
			oh1 := pipelines.NewObjectHasher()
			rp := RuntimeParameter{}
			writeRuntimeParameter(oh1, rp)

			oh2 := pipelines.NewObjectHasher()
			rp.ValueFrom = &ValueFrom{
				RunConfigurationRef: RunConfigurationRef{
					Name: apis.RandomString(),
				},
			}
			writeRuntimeParameter(oh2, rp)

			Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
		})

		Specify("artifact in valueFrom should change the hash", func() {
			oh1 := pipelines.NewObjectHasher()
			rp := RuntimeParameter{}
			writeRuntimeParameter(oh1, rp)

			oh2 := pipelines.NewObjectHasher()
			rp.ValueFrom = &ValueFrom{
				RunConfigurationRef: RunConfigurationRef{
					OutputArtifact: apis.RandomString(),
				},
			}
			writeRuntimeParameter(oh2, rp)

			Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
		})

		Specify("the same object should result in the same hash", func() {
			oh1 := pipelines.NewObjectHasher()
			rp := RuntimeParameter{
				Name: apis.RandomString(),
				Value: apis.RandomString(),
				ValueFrom: &ValueFrom{
					RunConfigurationRef: RunConfigurationRef{
						Name: apis.RandomString(),
						OutputArtifact: apis.RandomString(),
					},
				},
			}
			writeRuntimeParameter(oh1, rp)

			oh2 := pipelines.NewObjectHasher()
			writeRuntimeParameter(oh2, rp)

			Expect(oh1.Sum()).To(Equal(oh2.Sum()))
		})
	})
})
