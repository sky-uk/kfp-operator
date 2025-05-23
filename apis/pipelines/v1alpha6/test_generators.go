//go:build integration || decoupled || unit

package v1alpha6

import (
	"fmt"
	"math/rand"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	. "github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/argo/common"
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
		Spec:   RandomPipelineSpec(provider),
		Status: RandomStatus(provider),
	}
}

func RandomPipelineSpec(provider string) PipelineSpec {
	return PipelineSpec{
		Provider:      provider,
		Image:         fmt.Sprintf("%s:%s", RandomLowercaseString(), RandomShortHash()),
		TfxComponents: fmt.Sprintf("%s.%s", RandomLowercaseString(), RandomLowercaseString()),
		Env:           RandomNamedValues(),
		BeamArgs:      RandomNamedValues(),
	}
}

func RandomConditions() Conditions {
	return RandomList(RandomCondition)
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
	state := RandomSynchronizationState()
	return &RunConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: "default",
		},
		Spec: RandomRunConfigurationSpec(provider),
		Status: RunConfigurationStatus{
			SynchronizationState:     state,
			Provider:                 provider,
			ObservedPipelineVersion:  RandomString(),
			TriggeredPipelineVersion: RandomString(),
			Conditions:               Conditions{RandomSynchronizationStateCondition(state)},
		},
	}
}

func RandomRunConfigurationSpec(provider string) RunConfigurationSpec {
	return RunConfigurationSpec{
		Run:      RandomRunSpec(provider),
		Triggers: RandomTriggers(),
	}
}

func RandomTime() *metav1.Time {
	if rand.Intn(2) == 1 {
		return nil
	}
	min := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	randTime := time.Unix(rand.Int63n(max-min)+min, 0)
	return &metav1.Time{Time: randTime}
}

func RandomSchedule() Schedule {
	return Schedule{
		CronExpression: RandomString(),
		StartTime:      RandomTime(),
		EndTime:        RandomTime(),
	}
}

func RandomTriggers() Triggers {
	var onChange []OnChangeType
	if common.RandomInt64()%2 == 0 {
		onChange = []OnChangeType{OnChangeTypes.Pipeline}
	}

	return Triggers{
		Schedules: RandomList(RandomSchedule),
		OnChange:  onChange,
	}
}

func RandomScheduleTrigger() Triggers {
	return Triggers{Schedules: []Schedule{RandomSchedule()}}
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
		Spec:   RandomRunScheduleSpec(provider),
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

func RandomRunScheduleSpec(provider string) RunScheduleSpec {
	return RunScheduleSpec{
		Provider:          provider,
		Pipeline:          PipelineIdentifier{Name: RandomString(), Version: RandomString()},
		ExperimentName:    RandomString(),
		RuntimeParameters: RandomNamedValues(),
		Artifacts:         RandomList(RandomOutputArtifact),
		Schedule:          RandomSchedule(),
	}
}

func RandomRun(provider string) *Run {
	return &Run{
		ObjectMeta: metav1.ObjectMeta{
			Name:        RandomLowercaseString(),
			Namespace:   "default",
			Annotations: map[string]string{},
		},
		Spec: RandomRunSpec(provider),
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
				Name:           common.RandomNamespacedName(),
				OutputArtifact: RandomString(),
			},
		},
	}
}

func WithValueFrom(runSpec *RunSpec) {
	runSpec.RuntimeParameters = append(runSpec.RuntimeParameters, RandomList(RandomRunConfigurationRefRuntimeParameter)...)
}

func RandomRunSpec(provider string) RunSpec {
	return RunSpec{
		Provider:       provider,
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
		Spec:   RandomExperimentSpec(provider),
		Status: RandomStatus(provider),
	}
}

func RandomExperimentSpec(provider string) ExperimentSpec {
	return ExperimentSpec{
		Provider:    provider,
		Description: RandomString(),
	}
}

func RandomStatus(provider string) Status {
	state := RandomSynchronizationState()
	return Status{
		SynchronizationState: state,
		Version:              RandomString(),
		Provider: ProviderAndId{
			Name: provider,
			Id:   RandomString(),
		},
		ObservedGeneration: rand.Int63(),
		Conditions: Conditions{
			RandomSynchronizationStateCondition(state),
		},
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
		Status: RemoveProvider(RandomStatus(RandomString())),
	}
}

func RemoveProvider(status Status) Status {
	status.Provider = ProviderAndId{}
	return status
}

func RandomProviderSpec() ProviderSpec {
	randomParameters := make(map[string]*apiextensionsv1.JSON)
	for key := range RandomMap() {
		randomParameters[key] = &apiextensionsv1.JSON{Raw: []byte(`{"key1":"value1","key2":1234}`)}
	}

	return ProviderSpec{
		Image:               "kfp-operator-stub-provider",
		ServiceImage:        "service-image",
		ExecutionMode:       "none",
		ServiceAccount:      "default",
		DefaultBeamArgs:     RandomNamedValues(),
		PipelineRootStorage: RandomLowercaseString(),
		Parameters:          randomParameters,
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
		Status:          RandomStatus(RandomString()),
		NamespacedName:  RandomNamespacedName(),
		Kind:            RandomString(),
		ComputedVersion: RandomShortHash(),
	}
}
