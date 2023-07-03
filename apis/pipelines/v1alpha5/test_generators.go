//go:build integration || decoupled || unit
// +build integration decoupled unit

package v1alpha5

import (
	"fmt"
	. "github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/argo/common"
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
		Spec:   RandomPipelineSpec(),
		Status: RandomStatus(),
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
			SynchronizationState: RandomSynchronizationState(),
		},
	}
}

func RandomRunConfigurationSpec() RunConfigurationSpec {
	return RunConfigurationSpec{
		Run:      RandomRunSpec(),
		Triggers: RandomTriggers(),
	}
}

func RandomTriggers() Triggers {
	var onChange []OnChangeType
	if common.RandomInt64()%2 == 0 {
		onChange = []OnChangeType{OnChangeTypes.Pipeline}
	}

	return Triggers{
		Schedules: RandomList(RandomString),
		OnChange:  onChange,
	}
}

func RandomScheduleTrigger() Triggers {
	return Triggers{Schedules: []string{RandomString()}}
}

func RandomOnChangeTrigger() Triggers {
	return Triggers{OnChange: []OnChangeType{OnChangeTypes.Pipeline}}
}

func RandomRunSchedule() *RunSchedule {
	return &RunSchedule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: "default",
		},
		Spec:   RandomRunScheduleSpec(),
		Status: RandomStatus(),
	}
}

func RandomRunScheduleSpec() RunScheduleSpec {
	return RunScheduleSpec{
		RunSpec:  RandomRunSpec(),
		Schedule: RandomString(),
	}
}

func RandomRun() *Run {
	return &Run{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: "default",
		},
		Spec: RandomRunSpec(),
		Status: RunStatus{
			ObservedPipelineVersion: RandomString(),
			Status:                  RandomStatus(),
		},
	}
}

func WithValueFrom(runSpec *RunSpec) {
	runSpec.RuntimeParameters = append(runSpec.RuntimeParameters, RandomList(func() RuntimeParameter {
		return RuntimeParameter{
			Name:  RandomString(),
			ValueFrom: &ValueFrom{
				RunConfigurationRef: RunConfigurationRef{
					Name: RandomString(),
					OutputArtifact: RandomString(),
				},
			},
		}
	})...)
}

func RandomRunSpec() RunSpec {
	return RunSpec{
		Pipeline:       PipelineIdentifier{Name: RandomString(), Version: RandomString()},
		ExperimentName: RandomString(),
		RuntimeParameters: RandomList(func() RuntimeParameter {
			return RuntimeParameter{
				Name:  RandomString(),
				Value: RandomString(),
			}
		}),
		Artifacts: RandomList(func() OutputArtifact {
			return OutputArtifact{Name: RandomString(), Path: ArtifactPath{
				Locator: ArtifactLocator{
					Component: RandomString(),
					Artifact:  RandomString(),
					Index:     rand.Int(),
				},
				Filter: RandomString(),
			}}
		}),
	}
}

func RandomExperiment() *Experiment {
	return &Experiment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: "default",
		},
		Spec:   RandomExperimentSpec(),
		Status: RandomStatus(),
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
		ProviderId: ProviderAndId{
			Provider: RandomString(),
			Id:       RandomString(),
		},
		ObservedGeneration: rand.Int63(),
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

func (tr *TestResource) GetProvider() string {
	return tr.Status.ProviderId.Provider
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
