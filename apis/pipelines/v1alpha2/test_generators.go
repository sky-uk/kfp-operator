//go:build integration || decoupled || unit
// +build integration decoupled unit

package v1alpha2

import (
	"fmt"
	. "github.com/sky-uk/kfp-operator/apis"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RandomPipeline() *Pipeline {
	return &Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: "default",
		},
		Spec: RandomPipelineSpec(),
	}
}

func RandomPipelineSpec() PipelineSpec {
	return PipelineSpec{
		Image:         fmt.Sprintf("%s:%s", RandomLowercaseString(), RandomShortHash()),
		TfxComponents: fmt.Sprintf("%s.%s", RandomLowercaseString(), RandomLowercaseString()),
		Env:           RandomMap(),
		BeamArgs:      RandomMap(),
	}
}

func RandomRunConfiguration() *RunConfiguration {
	return &RunConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: "default",
		},
		Spec: RandomRunConfigurationSpec(),
		Status: RunConfigurationStatus{
			ObservedPipelineVersion: RandomString(),
		},
	}
}

func RandomRunConfigurationSpec() RunConfigurationSpec {
	return RunConfigurationSpec{
		Pipeline:          PipelineIdentifier{Name: RandomString(), Version: RandomString()},
		Schedule:          RandomString(),
		RuntimeParameters: RandomMap(),
	}
}

func RandomExperiment() *Experiment {
	return &Experiment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: "default",
		},
		Spec: RandomExperimentSpec(),
	}
}

func RandomExperimentSpec() ExperimentSpec {
	return ExperimentSpec{
		Description: RandomString(),
	}
}
