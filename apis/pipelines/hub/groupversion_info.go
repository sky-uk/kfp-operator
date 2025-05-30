// Package v1 contains API Schema definitions for the pipelines.kubeflow.org v1 API group
// +kubebuilder:object:generate=true
// +groupName=pipelines.kubeflow.org
package v1beta1

import (
	"github.com/sky-uk/kfp-operator/apis"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: apis.Group, Version: "v1beta1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
