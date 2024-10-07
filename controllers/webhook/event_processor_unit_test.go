//go:build unit

package webhook

import (
	"context"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func schemeWithCRDs() *runtime.Scheme {
	scheme := runtime.NewScheme()

	groupVersion := schema.GroupVersion{Group: "pipelines.kubeflow.org", Version: "v1alpha5"}
	scheme.AddKnownTypes(groupVersion, &v1alpha5.RunConfiguration{}, &v1alpha5.Run{})

	metav1.AddToGroupVersion(scheme, groupVersion)
	return scheme
}

func checkOutputArtifacts(outputArtifacts []v1alpha5.OutputArtifact, expectedArtifacts []v1alpha5.OutputArtifact) {
	if outputArtifacts == nil {
		Expect(expectedArtifacts).To(HaveLen(0))
	} else {
		Expect(outputArtifacts).To(Equal(expectedArtifacts))
	}
}

//func RandomRunCompletionEvent() kfpevents.RunCompletionEvent {
//	runName := kfpevents.RandomNamespacedName()
//	runConfigurationName := kfpevents.RandomNamespacedName()
//
//	return kfpevents.RunCompletionEvent{
//		Status:                kfpevents.RunCompletionStatus(kfpevents.RandomString()),
//		Provider:              kfpevents.RandomString(),
//		PipelineName:          kfpevents.RandomNamespacedName(),
//		RunName:               &runName,
//		RunConfigurationName:  &runConfigurationName,
//		RunId:                 kfpevents.RandomString(),
//		ServingModelArtifacts: RandomNonEmptyList(RandomArtifact),
//	}
//}

func RandomRunCompletionEventData() common.RunCompletionEventData {
	runName := common.RandomNamespacedName()
	runConfigurationName := common.RandomNamespacedName()

	return common.RunCompletionEventData{
		Status:                common.RunCompletionStatus(common.RandomString()),
		PipelineName:          common.NamespacedName{},
		RunConfigurationName:  &runConfigurationName,
		RunName:               &runName,
		RunId:                 common.RandomString(),
		ServingModelArtifacts: apis.RandomList(common.RandomArtifact),
		PipelineComponents:    apis.RandomList(common.RandomPipelineComponent),
		Provider:              common.RandomString(),
	}
}

//var _ = Context("ToRunCompletionEvent", func() {
//
//	When("", func() {
//		It("", func() {
//
//		})
//	})
//})

var _ = Context("filter", func() {
	basePipelineComponent := common.RandomPipelineComponent()
	baseOutputArtifact := v1alpha5.OutputArtifact{
		Name: "outputArtifactName",
		Path: v1alpha5.ArtifactPath{
			Locator: v1alpha5.ArtifactLocator{
				Component: basePipelineComponent.Name,
				Artifact:  basePipelineComponent.ComponentArtifacts[0].Name,
				Index:     0,
			},
			Filter: "pushed == 1",
		},
	}

	When("all fields match", func() {
		It("returns the matching artifact", func() {
			result := filter([]common.PipelineComponent{basePipelineComponent}, []v1alpha5.OutputArtifact{baseOutputArtifact})

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

			result := filter([]common.PipelineComponent{basePipelineComponent}, []v1alpha5.OutputArtifact{nonMatchingComponent})
			Expect(result).To(BeEmpty())
		})
	})

	When("artifact name is mismatched", func() {
		It("returns no artifact", func() {
			nonMatchingArtifactName := baseOutputArtifact
			nonMatchingArtifactName.Path.Locator.Artifact = "non matching artifact"

			result := filter([]common.PipelineComponent{basePipelineComponent}, []v1alpha5.OutputArtifact{nonMatchingArtifactName})
			Expect(result).To(BeEmpty())
		})
	})

	When("artifact is missing a uri", func() {
		It("returns no artifact", func() {
			missingArtifact := basePipelineComponent
			missingArtifact.ComponentArtifacts[0].Artifacts[0].Uri = ""

			result := filter([]common.PipelineComponent{missingArtifact}, []v1alpha5.OutputArtifact{baseOutputArtifact})
			Expect(result).To(BeEmpty())
		})
	})

	When("output artifact filter evaluates false on artifact metadata", func() {
		It("returns no artifact", func() {
			nonMatchingFilter := baseOutputArtifact
			nonMatchingFilter.Path.Filter = "pushed == 0"

			result := filter([]common.PipelineComponent{basePipelineComponent}, []v1alpha5.OutputArtifact{nonMatchingFilter})
			Expect(result).To(BeEmpty())
		})
	})

	When("output artifact locator index is more than component artifacts length", func() {
		It("returns no artifact", func() {
			invalidIndex := baseOutputArtifact
			invalidIndex.Path.Locator.Index = 999999999
			result := filter([]common.PipelineComponent{basePipelineComponent}, []v1alpha5.OutputArtifact{invalidIndex})
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
			rc := v1alpha5.RandomRunConfiguration()
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
			run := v1alpha5.RandomRun()
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
			rc := v1alpha5.RandomRunConfiguration()
			run := v1alpha5.RandomRun()
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
