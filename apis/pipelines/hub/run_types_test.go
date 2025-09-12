//go:build unit

package v1beta1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	"github.com/sky-uk/kfp-operator/pkg/common"
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

			run.Status.Dependencies.Pipeline.Version = "notempty"
			hash2 := run.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("All Parameters keys should change the hash", func() {
			run := Run{}
			hash1 := run.ComputeHash()

			run.Spec.Parameters = []Parameter{
				{Name: "a", Value: ""},
			}
			hash2 := run.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))

			run.Spec.Parameters = []Parameter{
				{Name: "b", Value: "notempty"},
			}
			hash3 := run.ComputeHash()

			Expect(hash2).NotTo(Equal(hash3))
		})

		Specify("The original object should not change", PropertyBased, func() {
			run := RandomRun(common.RandomNamespacedName())
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

	Context("cmpParameters", func() {
		unchanged := apis.RandomString()

		DescribeTable("cmpParameters",
			func(p1, p2 Parameter, expected bool) {
				Expect(cmpParameters(p1, p2)).To(Equal(expected))
			},
			Entry("first name less than",
				Parameter{Name: "A", Value: unchanged},
				Parameter{Name: "B", Value: unchanged},
				true),
			Entry("first name greater than",
				Parameter{Name: "B", Value: unchanged},
				Parameter{Name: "A", Value: unchanged},
				false),
			Entry("first value less than",
				Parameter{Name: unchanged, Value: "A"},
				Parameter{Name: unchanged, Value: "B"},
				true),
			Entry("first value greater than",
				Parameter{Name: unchanged, Value: "B"},
				Parameter{Name: unchanged, Value: "A"},
				false),
			Entry("same value",
				Parameter{Name: unchanged, Value: unchanged},
				Parameter{Name: unchanged, Value: unchanged},
				false),
			Entry("first runconfiguration name less than",
				Parameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: common.NamespacedName{Name: "A"}, OutputArtifact: unchanged}}},
				Parameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: common.NamespacedName{Name: "B"}, OutputArtifact: unchanged}}},
				true),
			Entry("first runconfiguration name greater than",
				Parameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: common.NamespacedName{Name: "B"}, OutputArtifact: unchanged}}},
				Parameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: common.NamespacedName{Name: "A"}, OutputArtifact: unchanged}}},
				false),
			Entry("first runconfiguration outputArtifact less than",
				Parameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: common.NamespacedName{Name: unchanged}, OutputArtifact: "A"}}},
				Parameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: common.NamespacedName{Name: unchanged}, OutputArtifact: "B"}}},
				true),
			Entry("first runconfiguration outputArtifact greater than",
				Parameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: common.NamespacedName{Name: unchanged}, OutputArtifact: "B"}}},
				Parameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: common.NamespacedName{Name: unchanged}, OutputArtifact: "A"}}},
				false),
			Entry("same name and outputArtifact but first optional",
				Parameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: common.NamespacedName{Name: unchanged}, OutputArtifact: unchanged, Optional: true}}},
				Parameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: common.NamespacedName{Name: unchanged}, OutputArtifact: unchanged}}},
				false),
			Entry("same name and outputArtifact but second optional",
				Parameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: common.NamespacedName{Name: unchanged}, OutputArtifact: unchanged}}},
				Parameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: common.NamespacedName{Name: unchanged}, OutputArtifact: unchanged, Optional: true}}},
				true),
			Entry("same valueFrom",
				Parameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: common.NamespacedName{Name: unchanged}, OutputArtifact: unchanged, Optional: true}}},
				Parameter{Name: unchanged, ValueFrom: &ValueFrom{RunConfigurationRef: RunConfigurationRef{Name: common.NamespacedName{Name: unchanged}, OutputArtifact: unchanged, Optional: true}}},
				false),
		)
	})

	Context("writeParameter", func() {
		Specify("name should change the hash", func() {
			oh1 := pipelines.NewObjectHasher()
			p := Parameter{}
			writeParameter(oh1, p)

			oh2 := pipelines.NewObjectHasher()
			p.Name = apis.RandomString()
			writeParameter(oh2, p)

			Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
		})

		Specify("value should change the hash", func() {
			oh1 := pipelines.NewObjectHasher()
			p := Parameter{}
			writeParameter(oh1, p)

			oh2 := pipelines.NewObjectHasher()
			p.Value = apis.RandomString()
			writeParameter(oh2, p)

			Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
		})

		Specify("name in valueFrom should change the hash", func() {
			oh1 := pipelines.NewObjectHasher()
			p := Parameter{}
			writeParameter(oh1, p)

			oh2 := pipelines.NewObjectHasher()
			p.ValueFrom = &ValueFrom{
				RunConfigurationRef: RunConfigurationRef{
					Name: common.RandomNamespacedName(),
				},
			}
			writeParameter(oh2, p)

			Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
		})

		Specify("artifact in valueFrom should change the hash", func() {
			oh1 := pipelines.NewObjectHasher()
			p := Parameter{}
			writeParameter(oh1, p)

			oh2 := pipelines.NewObjectHasher()
			p.ValueFrom = &ValueFrom{
				RunConfigurationRef: RunConfigurationRef{
					OutputArtifact: apis.RandomString(),
				},
			}
			writeParameter(oh2, p)

			Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
		})

		Specify("the same object should result in the same hash", func() {
			oh1 := pipelines.NewObjectHasher()
			p := Parameter{
				Name:  apis.RandomString(),
				Value: apis.RandomString(),
				ValueFrom: &ValueFrom{
					RunConfigurationRef: RunConfigurationRef{
						Name:           common.RandomNamespacedName(),
						OutputArtifact: apis.RandomString(),
					},
				},
			}
			writeParameter(oh1, p)

			oh2 := pipelines.NewObjectHasher()
			writeParameter(oh2, p)

			Expect(oh1.Sum()).To(Equal(oh2.Sum()))
		})
	})
})
var _ = Context("RunSpec", func() {
	Describe("ResolveParameters", func() {

		Specify("no ValueFrom", func() {
			expectedNamedValue := apis.RandomNamedValue()
			rs := RunSpec{
				Parameters: []Parameter{
					{
						Name:  expectedNamedValue.Name,
						Value: expectedNamedValue.Value,
					},
				},
			}

			namedValues, unresolvedOptionalParameters, err := rs.ResolveParameters(Dependencies{})
			Expect(err).NotTo(HaveOccurred())
			Expect(namedValues).To(ConsistOf(expectedNamedValue))
			Expect(unresolvedOptionalParameters).To(BeEmpty())
		})

		Specify("artifact not found in dependency but parameter is optional", func() {
			runConfigurationName := common.RandomNamespacedName()
			optionalParameter := Parameter{
				Name: apis.RandomString(),
				ValueFrom: &ValueFrom{
					RunConfigurationRef: RunConfigurationRef{
						Name:           runConfigurationName,
						OutputArtifact: apis.RandomString(),
						Optional:       true,
					},
				},
			}
			rs := RunSpec{Parameters: []Parameter{optionalParameter}}
			rcNamespacedName, err := runConfigurationName.String()
			Expect(err).NotTo(HaveOccurred())

			namedValues, unresolvedOptionalParameters, err := rs.ResolveParameters(Dependencies{
				RunConfigurations: map[string]RunReference{
					rcNamespacedName: {},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(namedValues).To(ConsistOf(apis.NamedValue{}))
			Expect(unresolvedOptionalParameters).To(Equal([]Parameter{optionalParameter}))
		})

		Specify("artifact not found in dependency and parameter is not optional", func() {
			runConfigurationName := common.RandomNamespacedName()
			rs := RunSpec{
				Parameters: []Parameter{
					{
						Name: apis.RandomString(),
						ValueFrom: &ValueFrom{
							RunConfigurationRef: RunConfigurationRef{
								Name:           runConfigurationName,
								OutputArtifact: apis.RandomString(),
							},
						},
					},
				},
			}
			rcNamespacedName, err := runConfigurationName.String()
			Expect(err).NotTo(HaveOccurred())

			_, unresolvedOptionalParameters, err := rs.ResolveParameters(Dependencies{
				RunConfigurations: map[string]RunReference{
					rcNamespacedName: {},
				},
			})

			Expect(unresolvedOptionalParameters).To(Equal([]Parameter{optionalParameter}))
			Expect(err).To(HaveOccurred())
		})

		Specify("dependency not found", func() {
			rs := RunSpec{
				Parameters: []Parameter{
					{
						Name: apis.RandomString(),
						ValueFrom: &ValueFrom{
							RunConfigurationRef: RunConfigurationRef{
								Name: common.RandomNamespacedName(),
							},
						},
					},
				},
			}

			_, _, err := rs.ResolveParameters(Dependencies{})
			Expect(err).To(HaveOccurred())
		})

		Specify("ValueFrom", func() {
			expectedNamedValue := apis.RandomNamedValue()
			runConfigurationName := common.RandomNamespacedName()

			artifact := apis.RandomString()

			rs := RunSpec{
				Parameters: []Parameter{
					{
						Name: expectedNamedValue.Name,
						ValueFrom: &ValueFrom{
							RunConfigurationRef: RunConfigurationRef{
								Name:           runConfigurationName,
								OutputArtifact: artifact,
							},
						},
					},
				},
			}

			rcNamespacedName, err := runConfigurationName.String()
			Expect(err).NotTo(HaveOccurred())

			namedValues, unresolvedOptionalParameters, err := rs.ResolveParameters(Dependencies{
				RunConfigurations: map[string]RunReference{
					rcNamespacedName: {
						Artifacts: []common.Artifact{
							{
								Name:     artifact,
								Location: expectedNamedValue.Value,
							},
						},
					},
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(namedValues).To(ConsistOf(expectedNamedValue))
			Expect(unresolvedOptionalParameters).To(BeEmpty())
		})
	})
})
