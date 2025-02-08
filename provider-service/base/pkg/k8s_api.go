package pkg

import (
	"github.com/sky-uk/kfp-operator/argo/common"
	"k8s.io/client-go/dynamic"
)

func NewK8sClient() (*K8sClient, error) {
	k8sConfig, err := common.K8sClientConfig()
	if err != nil {
		return nil, err
	}

	k8sClient, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	return &K8sClient{
		Client: k8sClient,
	}, nil
}

type K8sClient struct {
	Client dynamic.Interface
}
