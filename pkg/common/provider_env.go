package common

// Env var names derived from the Provider spec and injected into the
// provider-service container; the webhook also rejects them in podTemplateEnv.
const (
	ProviderNameEnvVar        = "PROVIDERNAME"
	PipelineRootStorageEnvVar = "PIPELINEROOTSTORAGE"
	ParametersEnvVarPrefix    = "PARAMETERS_"
)
