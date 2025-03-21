//go:build integration || decoupled || unit

package v1alpha5

import (
	"fmt"
	"math/rand"
	"time"

	. "github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/argo/common"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func RandomPipeline(provider string) *Pipeline {
	return &Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: "default",
		},
		Spec:   RandomPipelineSpec(),
		Status: RandomStatus(provider),
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

func RandomProvider() *Provider {
	return &Provider{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: "default",
		},
		Spec:   RandomProviderSpec(),
		Status: RandomProviderStatus(),
	}
}

func RandomProviderSpec() ProviderSpec {
	randomParameters := make(map[string]*apiextensionsv1.JSON)
	for key, _ := range RandomMap() {
		randomParameters[key] = &apiextensionsv1.JSON{Raw: []byte(`{"key1": "value1", "key2": 1234}`)}
	}

	return ProviderSpec{
		Image:               "kfp-operator-unused-old-provider-cli",
		ExecutionMode:       "none",
		ServiceAccount:      "default",
		DefaultBeamArgs:     RandomNamedValues(),
		PipelineRootStorage: RandomLowercaseString(),
		Parameters:          randomParameters,
	}
}

func RandomConditions() Conditions {
	return RandomList[metav1.Condition](RandomCondition)
}

func RandomCondition() metav1.Condition {
	return metav1.Condition{
		Type:               RandomLowercaseString(),
		Status:             RandomConditionStatus(),
		ObservedGeneration: common.RandomInt64(),
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             RandomLowercaseString(),
		Message:            RandomLowercaseString(),
	}
}

func RandomRunConfiguration(provider string) *RunConfiguration {
	return &RunConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: "default",
		},
		Spec: RandomRunConfigurationSpec(),
		Status: RunConfigurationStatus{
			SynchronizationState: RandomSynchronizationState(),
			Provider:             provider,
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

func RandomRunSchedule(provider string) *RunSchedule {
	return &RunSchedule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: "default",
		},
		Spec:   RandomRunScheduleSpec(),
		Status: RandomStatus(provider),
	}
}

func RandomOutputArtifact() OutputArtifact {
	return OutputArtifact{
		Name: RandomString(),
		Path: ArtifactPath{
			Locator: ArtifactLocator{
				Component: RandomString(),
				Artifact:  RandomString(),
				Index:     rand.Int(),
			},
			Filter: RandomString(),
		},
	}
}

func RandomRunScheduleSpec() RunScheduleSpec {
	return RunScheduleSpec{
		Pipeline:          PipelineIdentifier{Name: RandomString(), Version: RandomString()},
		ExperimentName:    RandomString(),
		RuntimeParameters: RandomNamedValues(),
		Artifacts:         RandomList(RandomOutputArtifact),
		Schedule:          RandomString(),
	}
}

func RandomRun(provider string) *Run {
	return &Run{
		ObjectMeta: metav1.ObjectMeta{
			Name:        RandomLowercaseString(),
			Namespace:   "default",
			Annotations: map[string]string{},
		},
		Spec: RandomRunSpec(),
		Status: RunStatus{
			ObservedPipelineVersion: RandomString(),
			Status:                  RandomStatus(provider),
		},
	}
}

func RandomRunConfigurationRefRuntimeParameter() RuntimeParameter {
	return RuntimeParameter{
		Name: RandomString(),
		ValueFrom: &ValueFrom{
			RunConfigurationRef: RunConfigurationRef{
				Name:           RandomString(),
				OutputArtifact: RandomString(),
			},
		},
	}
}

func WithValueFrom(runSpec *RunSpec) {
	runSpec.RuntimeParameters = append(runSpec.RuntimeParameters, RandomList(RandomRunConfigurationRefRuntimeParameter)...)
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

func RandomExperiment(provider string) *Experiment {
	return &Experiment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: "default",
		},
		Spec:   RandomExperimentSpec(),
		Status: RandomStatus(provider),
	}
}

func RandomExperimentSpec() ExperimentSpec {
	return ExperimentSpec{
		Description: RandomString(),
	}
}

func RandomStatus(provider string) Status {
	return Status{
		SynchronizationState: RandomSynchronizationState(),
		Version:              RandomString(),
		ProviderId: ProviderAndId{
			Provider: provider,
			Id:       RandomString(),
		},
		ObservedGeneration: rand.Int63(),
	}
}

// RandomProviderStatus differs from RandomStatus by not setting the provider
// because in practice it will never be populated.
func RandomProviderStatus() Status {
	return Status{
		SynchronizationState: RandomSynchronizationState(),
		Version:              RandomString(),
		ProviderId: ProviderAndId{
			Provider: "",
			Id:       "",
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
		Status:          RandomStatus(RandomString()),
		NamespacedName:  RandomNamespacedName(),
		Kind:            RandomString(),
		ComputedVersion: RandomShortHash(),
	}
}
