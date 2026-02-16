package config

type Config struct {
	Name                string     `yaml:"name"`
	Namespace           string     `yaml:"namespace"`
	PipelineRootStorage string     `yaml:"pipelineRootStorage"`
	Parameters          Parameters `yaml:"parameters"`
}

type Parameters struct {
	KfpNamespace             string `yaml:"kfpNamespace,omitempty"`
	RestKfpApiUrl            string `yaml:"restKfpApiUrl,omitempty"`
	GrpcMetadataStoreAddress string `yaml:"grpcMetadataStoreAddress,omitempty"`
	GrpcKfpApiAddress        string `yaml:"grpcKfpApiAddress,omitempty"`
}
