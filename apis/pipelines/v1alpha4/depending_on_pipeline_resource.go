package v1alpha4

// +kubebuilder:object:generate=false
type DependingOnPipelineResource interface {
	Resource
	GetObservedPipelineVersion() string
	SetObservedPipelineVersion(string)
}
