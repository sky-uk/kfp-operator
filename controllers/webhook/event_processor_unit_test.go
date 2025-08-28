//go:build unit

package webhook

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/pkg/common"
)

var _ = Context("ToRunCompletionEvent", func() {
	When("run configuration passed and no run passed", func() {
		It("returns RunCompletionEvent with filtered run configuration artifacts", func() {
			rc := pipelineshub.RandomRunConfiguration(common.RandomNamespacedName())
			rc.Spec.Run.Artifacts = []pipelineshub.OutputArtifact{{Name: "runconfig-artifact"}}

			runCompletionEventData := RandomRunCompletionEventData()
			runCompletionEventData.RunConfigurationName = &common.NamespacedName{
				Name:      rc.Name,
				Namespace: rc.Namespace,
			}
			runCompletionEventData.RunName = nil

			expectedArtifacts := []common.Artifact{{Name: "runconfig-artifact"}}

			result, err := ResourceArtifactsEventProcessor{filter: stubbedFilterFunc}.ToRunCompletionEvent(&runCompletionEventData, rc, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&common.RunCompletionEvent{
				Status:                runCompletionEventData.Status,
				PipelineName:          runCompletionEventData.PipelineName,
				RunConfigurationName:  runCompletionEventData.RunConfigurationName,
				RunName:               runCompletionEventData.RunName,
				RunId:                 runCompletionEventData.RunId,
				RunStartTime:          runCompletionEventData.RunStartTime,
				RunEndTime:            runCompletionEventData.RunEndTime,
				ServingModelArtifacts: runCompletionEventData.ServingModelArtifacts,
				Artifacts:             expectedArtifacts,
				Provider:              runCompletionEventData.Provider,
			}))

		})
	})

	When("run passed and no run configuration passed", func() {
		It("returns RunCompletionEvent with filtered run artifacts", func() {
			run := pipelineshub.RandomRun(common.RandomNamespacedName())
			run.Spec.Artifacts = []pipelineshub.OutputArtifact{{Name: "run-artifact"}}

			runCompletionEventData := RandomRunCompletionEventData()
			runCompletionEventData.RunName = &common.NamespacedName{
				Name:      run.Name,
				Namespace: run.Namespace,
			}
			runCompletionEventData.RunConfigurationName = nil

			expectedArtifacts := []common.Artifact{{Name: "run-artifact"}}

			result, err := ResourceArtifactsEventProcessor{filter: stubbedFilterFunc}.ToRunCompletionEvent(&runCompletionEventData, nil, run)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&common.RunCompletionEvent{
				Status:                runCompletionEventData.Status,
				PipelineName:          runCompletionEventData.PipelineName,
				RunConfigurationName:  runCompletionEventData.RunConfigurationName,
				RunName:               runCompletionEventData.RunName,
				RunId:                 runCompletionEventData.RunId,
				RunStartTime:          runCompletionEventData.RunStartTime,
				RunEndTime:            runCompletionEventData.RunEndTime,
				ServingModelArtifacts: runCompletionEventData.ServingModelArtifacts,
				Artifacts:             expectedArtifacts,
				Provider:              runCompletionEventData.Provider,
			}))
		})
	})

	When("both run configuration and run passed", func() {
		It("returns RunCompletionEvent with filtered run configuration artifacts", func() {
			rc := pipelineshub.RandomRunConfiguration(common.RandomNamespacedName())
			rc.Spec.Run.Artifacts = []pipelineshub.OutputArtifact{{Name: "runconfig-artifact"}}
			run := pipelineshub.RandomRun(common.RandomNamespacedName())
			run.Spec.Artifacts = []pipelineshub.OutputArtifact{{Name: "run-artifact"}}

			runCompletionEventData := RandomRunCompletionEventData()
			runCompletionEventData.RunConfigurationName = &common.NamespacedName{
				Name:      rc.Name,
				Namespace: rc.Namespace,
			}
			runCompletionEventData.RunName = &common.NamespacedName{
				Name:      run.Name,
				Namespace: run.Namespace,
			}

			expectedArtifacts := []common.Artifact{{Name: "runconfig-artifact"}}

			result, err := ResourceArtifactsEventProcessor{filter: stubbedFilterFunc}.ToRunCompletionEvent(&runCompletionEventData, rc, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&common.RunCompletionEvent{
				Status:                runCompletionEventData.Status,
				PipelineName:          runCompletionEventData.PipelineName,
				RunConfigurationName:  runCompletionEventData.RunConfigurationName,
				RunName:               runCompletionEventData.RunName,
				RunId:                 runCompletionEventData.RunId,
				RunStartTime:          runCompletionEventData.RunStartTime,
				RunEndTime:            runCompletionEventData.RunEndTime,
				ServingModelArtifacts: runCompletionEventData.ServingModelArtifacts,
				Artifacts:             expectedArtifacts,
				Provider:              runCompletionEventData.Provider,
			}))

		})
	})

	When("neither run configuration or run passed", func() {
		It("returns an error", func() {
			runCompletionEventData := RandomRunCompletionEventData()
			stubbedFilterFunc := func(_ []common.PipelineComponent, _ []pipelineshub.OutputArtifact) []common.Artifact {
				return []common.Artifact{}
			}

			_, err := ResourceArtifactsEventProcessor{filter: stubbedFilterFunc}.ToRunCompletionEvent(&runCompletionEventData, nil, nil)
			Expect(err).To(Equal(&InvalidEvent{err.Error()}))
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

func stubbedFilterFunc(_ []common.PipelineComponent, outputArtifacts []pipelineshub.OutputArtifact) []common.Artifact {
	filteredArtifacts := []common.Artifact{}
	for _, artifact := range outputArtifacts {
		filteredArtifacts = append(filteredArtifacts, common.Artifact{
			Name: artifact.Name,
		})
	}
	return filteredArtifacts
}
