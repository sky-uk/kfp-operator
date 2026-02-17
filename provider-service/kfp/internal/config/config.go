package config

import "github.com/sky-uk/kfp-operator/pkg/common"

type Config struct {
	ProviderName        common.NamespacedName `mapstructure:"providerName" yaml:"providerName"`
	PipelineRootStorage string                `mapstructure:"pipelineRootStorage" yaml:"pipelineRootStorage"`
	Parameters          Parameters            `mapstructure:"parameters" yaml:"parameters"`
}

type Parameters struct {
	KfpNamespace             string `mapstructure:"kfpNamespace" yaml:"kfpNamespace,omitempty"`
	RestKfpApiUrl            string `mapstructure:"restKfpApiUrl" yaml:"restKfpApiUrl,omitempty"`
	GrpcMetadataStoreAddress string `mapstructure:"grpcMetadataStoreAddress" yaml:"grpcMetadataStoreAddress,omitempty"`
	GrpcKfpApiAddress        string `mapstructure:"grpcKfpApiAddress" yaml:"grpcKfpApiAddress,omitempty"`
}
