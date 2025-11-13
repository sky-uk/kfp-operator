package label

import "github.com/sky-uk/kfp-operator/pkg/common/triggers"

const (
	ProviderName              = "provider-name"
	ProviderNamespace         = "provider-namespace"
	PipelineName              = "pipeline-name"
	PipelineNamespace         = "pipeline-namespace"
	PipelineVersion           = "pipeline-version"
	RunConfigurationName      = "runconfiguration-name"
	RunConfigurationNamespace = "runconfiguration-namespace"
	RunName                   = "run-name"
	RunNamespace              = "run-namespace"
	SchemaVersion             = "schema_version"
	SdkVersion                = "sdk_version"
)

var LabelKeys = []string{
	ProviderName,
	ProviderNamespace,
	PipelineName,
	PipelineNamespace,
	PipelineVersion,
	RunConfigurationName,
	RunConfigurationNamespace,
	RunName,
	RunNamespace,
	triggers.Type,
	triggers.Source,
	triggers.SourceNamespace,
}
