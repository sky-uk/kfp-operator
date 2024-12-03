package internal

import (
	"context"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg"
)

type KfpProviderConfig struct {
	Name       string     `yaml:"name"`
	Parameters Parameters `yaml:"parameters"`
}

func LoadProviderConfig(ctx context.Context, k8sClient pkg.K8sClient, providerName string, namespace string) (*KfpProviderConfig, error) {
	logger := common.LoggerFromContext(ctx)
	config := &KfpProviderConfig{
		Name: providerName,
	}

	if err := pkg.LoadProvider[KfpProviderConfig](ctx, k8sClient.Client, providerName, namespace, config); err != nil {
		logger.Error(err, "failed to load provider", "name", providerName, "namespace", namespace)
		return nil, err
	}
	return config, nil
}

type Parameters struct {
	KfpNamespace             string `yaml:"kfpNamespace,omitempty"`
	RestKfpApiUrl            string `yaml:"restKfpApiUrl,omitempty"`
	GrpcMetadataStoreAddress string `yaml:"grpcMetadataStoreAddress,omitempty"`
	GrpcKfpApiAddress        string `yaml:"grpcKfpApiAddress,omitempty"`
}
