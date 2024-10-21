//go:build unit

package webhook

import (
	"context"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func schemeWithCRDs() *runtime.Scheme {
	scheme := runtime.NewScheme()

	groupVersion := schema.GroupVersion{Group: "pipelines.kubeflow.org", Version: "v1alpha6"}
	scheme.AddKnownTypes(groupVersion, &pipelinesv1.RunConfiguration{}, &pipelinesv1.Run{})

	metav1.AddToGroupVersion(scheme, groupVersion)
	return scheme
}

func checkOutputArtifacts(outputArtifacts []pipelinesv1.OutputArtifact, expectedArtifacts []pipelinesv1.OutputArtifact) {
	if outputArtifacts == nil {
		Expect(expectedArtifacts).To(BeEmpty())
	} else {
		Expect(outputArtifacts).To(Equal(expectedArtifacts))
	}
}

var _ = Context("ToRunCompletionEvent", func() {
	var ctx = logr.NewContext(context.Background(), logr.Discard())

	When("given valid runCompletionEventData", func() {
		It("converts to a runCompletionEvent with filtered artifacts", func() {
			rc := pipelinesv1.RandomRunConfiguration()

			runCompletionEventData := RandomRunCompletionEventData()
			runCompletionEventData.RunConfigurationName = &common.NamespacedName{
				Name:      rc.Name,
				Namespace: rc.Namespace,
			}

			expectedArtifacts := apis.RandomNonEmptyList(common.RandomArtifact)

			stubbedFilterFunc := func(_ []common.PipelineComponent, _ []pipelinesv1.OutputArtifact) []common.Artifact {
				return expectedArtifacts
			}

			fakeClient := fake.NewClientBuilder().WithScheme(schemeWithCRDs()).WithObjects(rc).Build()

			result, err := ResourceArtifactsEventProcessor{client: fakeClient, filter: stubbedFilterFunc}.ToRunCompletionEvent(ctx, runCompletionEventData)
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
})

var _ = Context("filter", func() {
	basePipelineComponent := randomPipelineComponent()
	baseOutputArtifact := pipelinesv1.OutputArtifact{
		Name: "outputArtifactName",
		Path: pipelinesv1.ArtifactPath{
			Locator: pipelinesv1.ArtifactLocator{
				Component: basePipelineComponent.Name,
				Artifact:  basePipelineComponent.ComponentArtifacts[0].Name,
				Index:     0,
			},
			Filter: "pushed == 1",
		},
	}

	When("all fields match", func() {
		It("returns the matching artifact", func() {
			result := filterByResourceArtifacts([]common.PipelineComponent{basePipelineComponent}, []pipelinesv1.OutputArtifact{baseOutputArtifact})

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

			result := filterByResourceArtifacts([]common.PipelineComponent{basePipelineComponent}, []pipelinesv1.OutputArtifact{nonMatchingComponent})
			Expect(result).To(BeEmpty())
		})
	})

	When("artifact name is mismatched", func() {
		It("returns no artifact", func() {
			nonMatchingArtifactName := baseOutputArtifact
			nonMatchingArtifactName.Path.Locator.Artifact = "non matching artifact"

			result := filterByResourceArtifacts([]common.PipelineComponent{basePipelineComponent}, []pipelinesv1.OutputArtifact{nonMatchingArtifactName})
			Expect(result).To(BeEmpty())
		})
	})

	When("artifact is missing a uri", func() {
		It("returns no artifact", func() {
			missingArtifact := basePipelineComponent
			missingArtifact.ComponentArtifacts[0].Artifacts[0].Uri = ""

			result := filterByResourceArtifacts([]common.PipelineComponent{missingArtifact}, []pipelinesv1.OutputArtifact{baseOutputArtifact})
			Expect(result).To(BeEmpty())
		})
	})

	When("output artifact filter evaluates false on artifact metadata", func() {
		It("returns no artifact", func() {
			nonMatchingFilter := baseOutputArtifact
			nonMatchingFilter.Path.Filter = "pushed == 0"

			result := filterByResourceArtifacts([]common.PipelineComponent{basePipelineComponent}, []pipelinesv1.OutputArtifact{nonMatchingFilter})
			Expect(result).To(BeEmpty())
		})
	})

	When("output artifact locator index is more than component artifacts length", func() {
		It("returns no artifact", func() {
			invalidIndex := baseOutputArtifact
			invalidIndex.Path.Locator.Index = 999999999
			result := filterByResourceArtifacts([]common.PipelineComponent{basePipelineComponent}, []pipelinesv1.OutputArtifact{invalidIndex})
			Expect(result).To(BeEmpty())
		})
	})
})

var _ = Context("extractResourceArtifacts", func() {
	var ctx = logr.NewContext(context.Background(), logr.Discard())
	When("neither run configuration or run name namespace passed", func() {
		It("should return an error", func() {
			_, err := extractResourceArtifacts(ctx, fake.NewClientBuilder().Build(), nil, nil)
			Expect(err).To(HaveOccurred())
		})
	})

	When("run configuration passed but resource not available", func() {
		It("should return an error", func() {
			rcName := &common.NamespacedName{
				Namespace: "rc-namespace",
				Name:      "rc-name",
			}
			_, err := extractResourceArtifacts(ctx, fake.NewClientBuilder().WithScheme(schemeWithCRDs()).Build(), rcName, nil)
			Expect(err).To(HaveOccurred())
		})
	})

	When("run configuration passed and no run name namespace", func() {
		It("should return run configuration artifacts", func() {
			rc := pipelinesv1.RandomRunConfiguration()
			rcName := &common.NamespacedName{
				Namespace: rc.Namespace,
				Name:      rc.Name,
			}
			fakeClient := fake.NewClientBuilder().WithScheme(schemeWithCRDs()).WithObjects(rc).Build()
			outputArtifacts, err := extractResourceArtifacts(ctx, fakeClient, rcName, nil)
			Expect(err).NotTo(HaveOccurred())
			checkOutputArtifacts(outputArtifacts, rc.Spec.Run.Artifacts)
		})
	})

	When("run passed and no run configuration", func() {
		It("should return run artifacts", func() {
			run := pipelinesv1.RandomRun()
			rName := &common.NamespacedName{
				Namespace: run.Namespace,
				Name:      run.Name,
			}
			fakeClient := fake.NewClientBuilder().WithScheme(schemeWithCRDs()).WithObjects(run).Build()
			outputArtifacts, err := extractResourceArtifacts(ctx, fakeClient, nil, rName)
			Expect(err).NotTo(HaveOccurred())
			checkOutputArtifacts(outputArtifacts, run.Spec.Artifacts)
		})
	})

	When("both run configuration and run passed", func() {
		It("should return run configuration artifacts", func() {
			rc := pipelinesv1.RandomRunConfiguration()
			run := pipelinesv1.RandomRun()
			rName := &common.NamespacedName{
				Namespace: run.Namespace,
				Name:      run.Name,
			}
			rcName := &common.NamespacedName{
				Namespace: rc.Namespace,
				Name:      rc.Name,
			}

			fakeClient := fake.NewClientBuilder().WithScheme(schemeWithCRDs()).WithObjects(rc, run).Build()
			outputArtifacts, err := extractResourceArtifacts(ctx, fakeClient, rcName, rName)
			Expect(err).NotTo(HaveOccurred())

			checkOutputArtifacts(outputArtifacts, rc.Spec.Run.Artifacts)
		})
	})
})
