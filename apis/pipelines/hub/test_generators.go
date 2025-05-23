//go:build integration || decoupled || unit

package v1beta1

import (
	"encoding/json"
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

func RandomPipeline(provider common.NamespacedName) *Pipeline {
	return &Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: "default",
		},
		Spec:   RandomPipelineSpec(provider),
		Status: RandomStatus(provider),
	}
}

func AddTfxValues(pipelineSpec *PipelineSpec) {
	pipelineSpec.Framework.Name = "tfx"
	pipelineSpec.Framework.Parameters = make(map[string]*apiextensionsv1.JSON)
	component, _ := json.Marshal(RandomString())
	pipelineSpec.Framework.Parameters["components"] = &apiextensionsv1.JSON{Raw: component}

	beamArgs := []NamedValue{
		{Name: "key1", Value: "value1"},
		{Name: "key2", Value: "1234"},
	}

	beamArgsMarshalled, _ := json.Marshal(beamArgs)
	pipelineSpec.Framework.Parameters["beamArgs"] = &apiextensionsv1.JSON{Raw: beamArgsMarshalled}
}

func RandomPipelineSpec(provider common.NamespacedName) PipelineSpec {
	randParams := RandomMap()
	randomParameters := make(map[string]*apiextensionsv1.JSON)
	for key, value := range randParams {
		randomValue, _ := json.Marshal(value)
		a := apiextensionsv1.JSON{Raw: randomValue}
		randomParameters[key] = &a
	}

	return PipelineSpec{
		Provider: provider,
		Image:    fmt.Sprintf("%s:%s", RandomLowercaseString(), RandomShortHash()),
		Env:      RandomNamedValues(),
		Framework: PipelineFramework{
			Name:       RandomString(),
			Parameters: randomParameters,
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
		Status: RemoveProvider(RandomStatus(common.NamespacedName{})),
	}
}

// RemoveProvider removes the provider from the status
// This is because providers never have this field populated, nor should they,
// but they share the Status type with other resources that do use this field.
func RemoveProvider(status Status) Status {
	status.Provider = ProviderAndId{}
	return status
}

func RandomProviderSpec() ProviderSpec {
	randomParameters := make(map[string]*apiextensionsv1.JSON)
	for key := range RandomMap() {
		randomParameters[key] = &apiextensionsv1.JSON{Raw: []byte(`{"key1": "value1", "key2": 1234}`)}
	}

	return ProviderSpec{
		ServiceImage:        "service-image",
		ExecutionMode:       "none",
		ServiceAccount:      "default",
		PipelineRootStorage: RandomLowercaseString(),
		Frameworks:          RandomList(RandomFramework),
		Parameters:          randomParameters,
		AllowedNamespaces:   RandomList(RandomString),
	}
}

func RandomFramework() Framework {
	return Framework{
		Name:    RandomString(),
		Image:   RandomString(),
		Patches: []Patch{},
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

func RandomRunConfiguration(provider common.NamespacedName) *RunConfiguration {
	return &RunConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: "default",
		},
		Spec: RandomRunConfigurationSpec(provider),
		Status: RunConfigurationStatus{
			Provider: provider,
			Conditions: Conditions{
				RandomSynchronizationStateCondition(RandomSynchronizationState()),
			},
		},
	}
}

func RandomRunConfigurationSpec(provider common.NamespacedName) RunConfigurationSpec {
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

func RandomRunSchedule(provider common.NamespacedName) *RunSchedule {
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

func RandomRunScheduleSpec(provider common.NamespacedName) RunScheduleSpec {
	return RunScheduleSpec{
		Provider:       provider,
		Pipeline:       PipelineIdentifier{Name: RandomString(), Version: RandomString()},
		ExperimentName: RandomString(),
		Parameters:     RandomNamedValues(),
		Artifacts:      RandomList(RandomOutputArtifact),
		Schedule:       RandomSchedule(),
	}
}

func RandomRun(provider common.NamespacedName) *Run {
	return &Run{
		ObjectMeta: metav1.ObjectMeta{
			Name:        RandomLowercaseString(),
			Namespace:   "default",
			Annotations: map[string]string{},
		},
		Spec: RandomRunSpec(provider),
		Status: RunStatus{
			Status:       RandomStatus(provider),
			Dependencies: RandomDependencies(),
		},
	}
}

func RandomRunConfigurationRefParameter() Parameter {
	return Parameter{
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
	runSpec.Parameters = append(runSpec.Parameters, RandomList(RandomRunConfigurationRefParameter)...)
}

func RandomRunSpec(provider common.NamespacedName) RunSpec {
	return RunSpec{
		Provider:       provider,
		Pipeline:       PipelineIdentifier{Name: RandomString(), Version: RandomString()},
		ExperimentName: RandomString(),
		Parameters: RandomList(func() Parameter {
			return Parameter{
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

func RandomExperiment(provider common.NamespacedName) *Experiment {
	return &Experiment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: "default",
		},
		Spec:   RandomExperimentSpec(provider),
		Status: RandomStatus(provider),
	}
}

func RandomExperimentSpec(provider common.NamespacedName) ExperimentSpec {
	return ExperimentSpec{
		Provider:    provider,
		Description: RandomString(),
	}
}

func RandomStatus(provider common.NamespacedName) Status {
	return Status{
		Version: RandomString(),
		Provider: ProviderAndId{
			Name: provider,
			Id:   RandomString(),
		},
		ObservedGeneration: rand.Int63(),
		Conditions: Conditions{
			RandomSynchronizationStateCondition(
				RandomSynchronizationState(),
			),
		},
	}
}

func RandomDependencies() Dependencies {
	return Dependencies{
		Pipeline: ObservedPipeline{
			Version: RandomString(),
		},
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
		Status:          RandomStatus(common.RandomNamespacedName()),
		NamespacedName:  RandomNamespacedName(),
		Kind:            RandomString(),
		ComputedVersion: RandomShortHash(),
	}
}
