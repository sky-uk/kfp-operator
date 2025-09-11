package config

type Config struct {
	Name                string     `yaml:"name"`
	PipelineRootStorage string     `yaml:"pipelineRootStorage"`
	Parameters          Parameters `yaml:"parameters"`
}

type Parameters struct {
	KfpNamespace             string `yaml:"kfpNamespace,omitempty"`
	RestKfpApiUrl            string `yaml:"restKfpApiUrl,omitempty"`
	GrpcMetadataStoreAddress string `yaml:"grpcMetadataStoreAddress,omitempty"`
	GrpcKfpApiAddress        string `yaml:"grpcKfpApiAddress,omitempty"`
}
