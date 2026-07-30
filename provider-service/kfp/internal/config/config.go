package config

import "github.com/sky-uk/kfp-operator/pkg/common"

// BearerTokenPath is the fixed path at which the KFP provider-service reads the
// projected ServiceAccount token when running in multi-user mode. Provider
// owners must mount their projected token at this path via the Provider CR's
// podTemplateVolumes/podTemplateVolumeMounts.
const BearerTokenPath = "/var/run/secrets/kfp/token"

type Config struct {
	ProviderName        common.NamespacedName `mapstructure:"providerName" yaml:"providerName"`
	PipelineRootStorage string                `mapstructure:"pipelineRootStorage" yaml:"pipelineRootStorage"`
	Parameters          Parameters            `mapstructure:"parameters" yaml:"parameters"`
}

type Parameters struct {
	KfpNamespace             string `mapstructure:"kfpNamespace" yaml:"kfpNamespace,omitempty"`
	KfpMultiUserMode         bool   `mapstructure:"kfpMultiUserMode" yaml:"kfpMultiUserMode,omitempty"`
	RestKfpApiUrl            string `mapstructure:"restKfpApiUrl" yaml:"restKfpApiUrl,omitempty"`
	GrpcMetadataStoreAddress string `mapstructure:"grpcMetadataStoreAddress" yaml:"grpcMetadataStoreAddress,omitempty"`
	GrpcKfpApiAddress        string `mapstructure:"grpcKfpApiAddress" yaml:"grpcKfpApiAddress,omitempty"`
	DefaultExperiment        string `mapstructure:"defaultExperiment" yaml:"defaultExperiment,omitempty"`
}
