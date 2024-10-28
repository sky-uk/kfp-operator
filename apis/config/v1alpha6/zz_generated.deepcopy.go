//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha6

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Endpoint) DeepCopyInto(out *Endpoint) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Endpoint.
func (in *Endpoint) DeepCopy() *Endpoint {
	if in == nil {
		return nil
	}
	out := new(Endpoint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KfpControllerConfig) DeepCopyInto(out *KfpControllerConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.ControllerManagerConfigurationSpec.DeepCopyInto(&out.ControllerManagerConfigurationSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KfpControllerConfig.
func (in *KfpControllerConfig) DeepCopy() *KfpControllerConfig {
	if in == nil {
		return nil
	}
	out := new(KfpControllerConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KfpControllerConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KfpControllerConfigSpec) DeepCopyInto(out *KfpControllerConfigSpec) {
	*out = *in
	if in.RunCompletionTTL != nil {
		in, out := &in.RunCompletionTTL, &out.RunCompletionTTL
		*out = new(v1.Duration)
		**out = **in
	}
	in.RunCompletionFeed.DeepCopyInto(&out.RunCompletionFeed)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KfpControllerConfigSpec.
func (in *KfpControllerConfigSpec) DeepCopy() *KfpControllerConfigSpec {
	if in == nil {
		return nil
	}
	out := new(KfpControllerConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceConfiguration) DeepCopyInto(out *ServiceConfiguration) {
	*out = *in
	if in.Endpoints != nil {
		in, out := &in.Endpoints, &out.Endpoints
		*out = make([]Endpoint, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceConfiguration.
func (in *ServiceConfiguration) DeepCopy() *ServiceConfiguration {
	if in == nil {
		return nil
	}
	out := new(ServiceConfiguration)
	in.DeepCopyInto(out)
	return out
}