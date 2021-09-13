package workflows

import (
	"fmt"

	"math/rand"

	pipelinesv1 "github.com/sky-uk/kfp-operator/api/v1"
	"github.com/thanhpk/randstr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func randomLowercaseString() string {
	return randstr.String(rand.Intn(20), "0123456789abcdefghijklmnopqrstuvwxyz")
}

func randomShortHash() string {
	return randstr.String(7, "0123456789abcdef")
}

func randomString() string {
	return randstr.String(rand.Intn(20))
}

func randomMap() map[string]string {
	size := rand.Intn(5)

	rMap := make(map[string]string, size)
	for i := 1; i <= size; i++ {
		rMap[randomString()] = randomString()
	}

	return rMap
}

func RandomPipeline() *pipelinesv1.Pipeline {
	return &pipelinesv1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      randomLowercaseString(),
			Namespace: randomLowercaseString(),
		},
		Spec: pipelinesv1.PipelineSpec{
			Image:         fmt.Sprintf("%s:%s", randomLowercaseString(), randomShortHash()),
			TfxComponents: fmt.Sprintf("%s.%s", randomLowercaseString(), randomLowercaseString()),
			Env:           randomMap(),
		},
	}
}
