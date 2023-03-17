package v1alpha4

// +kubebuilder:object:generate=false
type DependingOnPipelineResource interface {
	Resource
	GetPipeline() PipelineIdentifier
	GetObservedPipelineVersion() string
	SetObservedPipelineVersion(string)
}
