package config

type KfpProviderConfig struct {
	Name       string     `yaml:"name"`
	Parameters Parameters `yaml:"parameters"`
}

type Parameters struct {
	KfpNamespace             string `yaml:"kfpNamespace,omitempty"`
	RestKfpApiUrl            string `yaml:"restKfpApiUrl,omitempty"`
	GrpcMetadataStoreAddress string `yaml:"grpcMetadataStoreAddress,omitempty"`
	GrpcKfpApiAddress        string `yaml:"grpcKfpApiAddress,omitempty"`
}
