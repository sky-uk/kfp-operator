package base

import (
	"context"
	"encoding/json"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

var ProviderGVR = pipelinesv1.GroupVersion.WithResource("providers")

func LoadProvider[Config any](ctx context.Context, k8sClient dynamic.Interface, provider string, namespace string, config *Config) error {
	providerConfig, err := k8sClient.Resource(ProviderGVR).Namespace(namespace).Get(ctx, provider, metav1.GetOptions{}, "")
	if err != nil {
		return err
	}

	providerCR := pipelinesv1.Provider{}

	if err = k8runtime.DefaultUnstructuredConverter.FromUnstructured(providerConfig.UnstructuredContent(), &providerCR); err != nil {
		return err
	}

	spec := providerCR.Spec

	specMarshalled, err := json.Marshal(spec)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(specMarshalled, &config); err != nil {
		return err
	}

	return nil
}

func CreateK8sClient() (dynamic.Interface, error) {
	k8sConfig, err := common.K8sClientConfig()
	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(k8sConfig)
}

type K8sApi struct {
	K8sClient dynamic.Interface
}
