//go:build integration || decoupled || unit
// +build integration decoupled unit

package pipelines

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"math/rand"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha1"
	"github.com/thanhpk/randstr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func RandomLowercaseString() string {
	return randstr.String(rand.Intn(20)+1, "0123456789abcdefghijklmnopqrstuvwxyz")
}

func RandomShortHash() string {
	return randstr.String(7, "0123456789abcdef")
}

func RandomString() string {
	return randstr.String(rand.Intn(20) + 1)
}

func RandomMap() map[string]string {
	size := rand.Intn(5)

	rMap := make(map[string]string, size)
	for i := 1; i <= size; i++ {
		rMap[RandomString()] = RandomString()
	}

	return rMap
}

type TestResource struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	NamespacedName types.NamespacedName

	Kind   string
	Status pipelinesv1.Status
}

func (tr *TestResource) GetStatus() pipelinesv1.Status {
	return tr.GetStatus()
}

func (tr *TestResource) SetStatus(status pipelinesv1.Status) {
	tr.Status = status
}

func (tr *TestResource) DeepCopyObject() runtime.Object {
	return tr
}

func (tr *TestResource) GetNamespacedName() types.NamespacedName {
	return tr.NamespacedName
}

func (tr *TestResource) GetKind() string {
	return tr.Kind
}

func RandomResource() Resource {
	return &TestResource{
		Status:         RandomStatus(),
		NamespacedName: RandomNamespacedName(),
		Kind:           RandomString(),
	}
}

func RandomPipeline() *pipelinesv1.Pipeline {
	return &pipelinesv1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: RandomLowercaseString(),
		},
		Spec: RandomPipelineSpec(),
	}
}

func RandomPipelineSpec() pipelinesv1.PipelineSpec {
	return pipelinesv1.PipelineSpec{
		Image:         fmt.Sprintf("%s:%s", RandomLowercaseString(), RandomShortHash()),
		TfxComponents: fmt.Sprintf("%s.%s", RandomLowercaseString(), RandomLowercaseString()),
		Env:           RandomMap(),
		BeamArgs:      RandomMap(),
	}
}

func RandomRunConfiguration() *pipelinesv1.RunConfiguration {
	return &pipelinesv1.RunConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: RandomLowercaseString(),
		},
		Spec: RandomRunConfigurationSpec(),
	}
}

func RandomRunConfigurationSpec() pipelinesv1.RunConfigurationSpec {
	return pipelinesv1.RunConfigurationSpec{
		PipelineName:      RandomString(),
		Schedule:          RandomString(),
		RuntimeParameters: RandomMap(),
	}
}

func RandomExperiment() *pipelinesv1.Experiment {
	return &pipelinesv1.Experiment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RandomLowercaseString(),
			Namespace: RandomLowercaseString(),
		},
		Spec: RandomExperimentSpec(),
	}
}

func RandomExperimentSpec() pipelinesv1.ExperimentSpec {
	return pipelinesv1.ExperimentSpec{
		Description: RandomString(),
	}
}

func RandomSynchronizationState() pipelinesv1.SynchronizationState {
	synchronizationStates := []pipelinesv1.SynchronizationState{
		"",
		pipelinesv1.Creating,
		pipelinesv1.Succeeded,
		pipelinesv1.Updating,
		pipelinesv1.Deleting,
		pipelinesv1.Deleted,
		pipelinesv1.Failed,
	}

	return synchronizationStates[rand.Intn(len(synchronizationStates))]
}

func RandomStatus() pipelinesv1.Status {
	return pipelinesv1.Status{
		SynchronizationState: RandomSynchronizationState(),
		Version:              RandomString(),
		KfpId:                RandomString(),
		ObservedGeneration:   rand.Int63(),
	}
}

func RandomNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      RandomString(),
		Namespace: RandomString(),
	}
}
