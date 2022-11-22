//go:build integration || decoupled || unit
// +build integration decoupled unit

package v1alpha4

import (
	"fmt"
	. "github.com/sky-uk/kfp-operator/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"math/rand"
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
		Env:           RandomNamedValues(),
		BeamArgs:      RandomNamedValues(),
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
		RuntimeParameters: RandomNamedValues(),
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

func RandomStatus() Status {
	return Status{
		SynchronizationState: RandomSynchronizationState(),
		Version:              RandomString(),
		ProviderId:           RandomString(),
		ObservedGeneration:   rand.Int63(),
	}
}

type TestResource struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	NamespacedName types.NamespacedName

	Kind            string
	Status          Status
	ComputedVersion string
}

func (tr *TestResource) GetStatus() Status {
	return tr.Status
}

func (tr *TestResource) SetStatus(status Status) {
	tr.Status = status
}

func (tr *TestResource) DeepCopyObject() runtime.Object {
	return tr
}

func (tr *TestResource) GetNamespacedName() types.NamespacedName {
	return tr.NamespacedName
}

func (tr *TestResource) ComputeVersion() string {
	return tr.ComputedVersion
}

func (tr *TestResource) SetComputedVersion(version string) {
	tr.ComputedVersion = version
}

func (tr *TestResource) GetKind() string {
	return tr.Kind
}

func RandomResource() *TestResource {
	return &TestResource{
		Status:          RandomStatus(),
		NamespacedName:  RandomNamespacedName(),
		Kind:            RandomString(),
		ComputedVersion: RandomShortHash(),
	}
}