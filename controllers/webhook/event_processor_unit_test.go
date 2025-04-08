//go:build unit

package webhook

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
)

func checkOutputArtifacts(outputArtifacts []pipelineshub.OutputArtifact, expectedArtifacts []pipelineshub.OutputArtifact) {
	if outputArtifacts == nil {
		Expect(expectedArtifacts).To(BeEmpty())
	} else {
		Expect(outputArtifacts).To(Equal(expectedArtifacts))
	}
}

var _ = Context("ToRunCompletionEvent", func() {
	When("run configuration passed and no run passed", func() {
		It("returns RunCompletionEvent with filtered run configuration artifacts", func() {
			rc := pipelineshub.RandomRunConfiguration(common.RandomNamespacedName())

			runCompletionEventData := RandomRunCompletionEventData()
			runCompletionEventData.RunConfigurationName = &common.NamespacedName{
				Name:      rc.Name,
				Namespace: rc.Namespace,
			}

			// TODO: The expected artefacts should be the filtered output artifacts of the RC instead of randoms
			expectedArtifacts := apis.RandomNonEmptyList(common.RandomArtifact)

			stubbedFilterFunc := func(_ []common.PipelineComponent, _ []pipelineshub.OutputArtifact) []common.Artifact {
				return expectedArtifacts
			}

			result, err := ResourceArtifactsEventProcessor{filter: stubbedFilterFunc}.ToRunCompletionEvent(&runCompletionEventData, rc, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&common.RunCompletionEvent{
				Status:                runCompletionEventData.Status,
				PipelineName:          runCompletionEventData.PipelineName,
				RunConfigurationName:  runCompletionEventData.RunConfigurationName,
				RunName:               runCompletionEventData.RunName,
				RunId:                 runCompletionEventData.RunId,
				ServingModelArtifacts: runCompletionEventData.ServingModelArtifacts,
				Artifacts:             expectedArtifacts,
				Provider:              runCompletionEventData.Provider,
			}))

		})
	})

	When("run passed and no run configuration passed", func() {
		It("returns RunCompletionEvent with filtered run artifacts", func() {
			run := pipelineshub.RandomRun(common.RandomNamespacedName())

			runCompletionEventData := RandomRunCompletionEventData()
			runCompletionEventData.RunName = &common.NamespacedName{
				Name:      run.Name,
				Namespace: run.Namespace,
			}

			// TODO: The expected artefacts should be the filtered output artifacts of the Run instead of randoms
			expectedArtifacts := apis.RandomNonEmptyList(common.RandomArtifact)

			stubbedFilterFunc := func(_ []common.PipelineComponent, _ []pipelineshub.OutputArtifact) []common.Artifact {
				return expectedArtifacts
			}

			result, err := ResourceArtifactsEventProcessor{filter: stubbedFilterFunc}.ToRunCompletionEvent(&runCompletionEventData, nil, run)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&common.RunCompletionEvent{
				Status:                runCompletionEventData.Status,
				PipelineName:          runCompletionEventData.PipelineName,
				RunConfigurationName:  runCompletionEventData.RunConfigurationName,
				RunName:               runCompletionEventData.RunName,
				RunId:                 runCompletionEventData.RunId,
				ServingModelArtifacts: runCompletionEventData.ServingModelArtifacts,
				Artifacts:             expectedArtifacts,
				Provider:              runCompletionEventData.Provider,
			}))
		})
	})

	When("both run configuration and run passed", func() {
		It("returns RunCompletionEvent with filtered run configuration artifacts", func() {
			rc := pipelineshub.RandomRunConfiguration(common.RandomNamespacedName())
			run := pipelineshub.RandomRun(common.RandomNamespacedName())

			runCompletionEventData := RandomRunCompletionEventData()
			runCompletionEventData.RunConfigurationName = &common.NamespacedName{
				Name:      rc.Name,
				Namespace: rc.Namespace,
			}
			runCompletionEventData.RunName = &common.NamespacedName{
				Name:      run.Name,
				Namespace: run.Namespace,
			}

			// TODO: The expected artefacts should be the filtered output artifacts of the RC instead of randoms
			expectedArtifacts := apis.RandomNonEmptyList(common.RandomArtifact)

			stubbedFilterFunc := func(_ []common.PipelineComponent, _ []pipelineshub.OutputArtifact) []common.Artifact {
				return expectedArtifacts
			}

			result, err := ResourceArtifactsEventProcessor{filter: stubbedFilterFunc}.ToRunCompletionEvent(&runCompletionEventData, rc, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&common.RunCompletionEvent{
				Status:                runCompletionEventData.Status,
				PipelineName:          runCompletionEventData.PipelineName,
				RunConfigurationName:  runCompletionEventData.RunConfigurationName,
				RunName:               runCompletionEventData.RunName,
				RunId:                 runCompletionEventData.RunId,
				ServingModelArtifacts: runCompletionEventData.ServingModelArtifacts,
				Artifacts:             expectedArtifacts,
				Provider:              runCompletionEventData.Provider,
			}))

		})
	})

	When("neither run configuration nor run passed", func() {
		It("returns an error", func() {
			runCompletionEventData := RandomRunCompletionEventData()
			stubbedFilterFunc := func(_ []common.PipelineComponent, _ []pipelineshub.OutputArtifact) []common.Artifact {
				return []common.Artifact{}
			}

			_, err := ResourceArtifactsEventProcessor{filter: stubbedFilterFunc}.ToRunCompletionEvent(&runCompletionEventData, nil, nil)
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Context("filter", func() {
	basePipelineComponent := randomPipelineComponent()
	baseOutputArtifact := pipelineshub.OutputArtifact{
		Name: "outputArtifactName",
		Path: pipelineshub.ArtifactPath{
			Locator: pipelineshub.ArtifactLocator{
				Component: basePipelineComponent.Name,
				Artifact:  basePipelineComponent.ComponentArtifacts[0].Name,
				Index:     0,
			},
			Filter: "pushed == 1",
		},
	}

	When("all fields match", func() {
		It("returns the matching artifact", func() {
			result := filterByResourceArtifacts([]common.PipelineComponent{basePipelineComponent}, []pipelineshub.OutputArtifact{baseOutputArtifact})

			Expect(result).To(Equal([]common.Artifact{
				{
					Name:     baseOutputArtifact.Name,
					Location: basePipelineComponent.ComponentArtifacts[0].Artifacts[0].Uri,
				},
			}))
		})
	})

	When("component name is mismatched", func() {
		It("returns no artifact", func() {
			nonMatchingComponent := baseOutputArtifact
			nonMatchingComponent.Path.Locator.Component = "non matching component"

			result := filterByResourceArtifacts([]common.PipelineComponent{basePipelineComponent}, []pipelineshub.OutputArtifact{nonMatchingComponent})
			Expect(result).To(BeEmpty())
		})
	})

	When("artifact name is mismatched", func() {
		It("returns no artifact", func() {
			nonMatchingArtifactName := baseOutputArtifact
			nonMatchingArtifactName.Path.Locator.Artifact = "non matching artifact"

			result := filterByResourceArtifacts([]common.PipelineComponent{basePipelineComponent}, []pipelineshub.OutputArtifact{nonMatchingArtifactName})
			Expect(result).To(BeEmpty())
		})
	})

	When("artifact is missing a uri", func() {
		It("returns no artifact", func() {
			missingArtifact := basePipelineComponent
			missingArtifact.ComponentArtifacts[0].Artifacts[0].Uri = ""

			result := filterByResourceArtifacts([]common.PipelineComponent{missingArtifact}, []pipelineshub.OutputArtifact{baseOutputArtifact})
			Expect(result).To(BeEmpty())
		})
	})

	When("output artifact filter evaluates false on artifact metadata", func() {
		It("returns no artifact", func() {
			nonMatchingFilter := baseOutputArtifact
			nonMatchingFilter.Path.Filter = "pushed == 0"

			result := filterByResourceArtifacts([]common.PipelineComponent{basePipelineComponent}, []pipelineshub.OutputArtifact{nonMatchingFilter})
			Expect(result).To(BeEmpty())
		})
	})

	When("output artifact locator index is more than component artifacts length", func() {
		It("returns no artifact", func() {
			invalidIndex := baseOutputArtifact
			invalidIndex.Path.Locator.Index = 999999999
			result := filterByResourceArtifacts([]common.PipelineComponent{basePipelineComponent}, []pipelineshub.OutputArtifact{invalidIndex})
			Expect(result).To(BeEmpty())
		})
	})
})
