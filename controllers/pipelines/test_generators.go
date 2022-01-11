package pipelines

import (
	"fmt"

	"math/rand"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"github.com/thanhpk/randstr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RandomLowercaseString() string {
	return randstr.String(rand.Intn(20)+1, "0123456789abcdefghijklmnopqrstuvwxyz")
}

func RandomShortHash() string {
	return randstr.String(7, "0123456789abcdef")
}

func RandomString() string {
	return randstr.String(rand.Intn(20))
}

func RandomMap() map[string]string {
	size := rand.Intn(5)

	rMap := make(map[string]string, size)
	for i := 1; i <= size; i++ {
		rMap[RandomString()] = RandomString()
	}

	return rMap
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
		Description:      RandomString(),
	}
}
