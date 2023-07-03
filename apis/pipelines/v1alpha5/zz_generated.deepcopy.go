//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/argo/common"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArtifactLocator) DeepCopyInto(out *ArtifactLocator) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArtifactLocator.
func (in *ArtifactLocator) DeepCopy() *ArtifactLocator {
	if in == nil {
		return nil
	}
	out := new(ArtifactLocator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArtifactPath) DeepCopyInto(out *ArtifactPath) {
	*out = *in
	out.Locator = in.Locator
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArtifactPath.
func (in *ArtifactPath) DeepCopy() *ArtifactPath {
	if in == nil {
		return nil
	}
	out := new(ArtifactPath)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Dependencies) DeepCopyInto(out *Dependencies) {
	*out = *in
	if in.RunConfigurations != nil {
		in, out := &in.RunConfigurations, &out.RunConfigurations
		*out = make(map[string]RunReference, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Dependencies.
func (in *Dependencies) DeepCopy() *Dependencies {
	if in == nil {
		return nil
	}
	out := new(Dependencies)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Experiment) DeepCopyInto(out *Experiment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Experiment.
func (in *Experiment) DeepCopy() *Experiment {
	if in == nil {
		return nil
	}
	out := new(Experiment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Experiment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExperimentList) DeepCopyInto(out *ExperimentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Experiment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExperimentList.
func (in *ExperimentList) DeepCopy() *ExperimentList {
	if in == nil {
		return nil
	}
	out := new(ExperimentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ExperimentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExperimentSpec) DeepCopyInto(out *ExperimentSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExperimentSpec.
func (in *ExperimentSpec) DeepCopy() *ExperimentSpec {
	if in == nil {
		return nil
	}
	out := new(ExperimentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LatestRuns) DeepCopyInto(out *LatestRuns) {
	*out = *in
	in.Succeeded.DeepCopyInto(&out.Succeeded)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LatestRuns.
func (in *LatestRuns) DeepCopy() *LatestRuns {
	if in == nil {
		return nil
	}
	out := new(LatestRuns)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OutputArtifact) DeepCopyInto(out *OutputArtifact) {
	*out = *in
	out.Path = in.Path
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OutputArtifact.
func (in *OutputArtifact) DeepCopy() *OutputArtifact {
	if in == nil {
		return nil
	}
	out := new(OutputArtifact)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Pipeline) DeepCopyInto(out *Pipeline) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Pipeline.
func (in *Pipeline) DeepCopy() *Pipeline {
	if in == nil {
		return nil
	}
	out := new(Pipeline)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Pipeline) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PipelineIdentifier) DeepCopyInto(out *PipelineIdentifier) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PipelineIdentifier.
func (in *PipelineIdentifier) DeepCopy() *PipelineIdentifier {
	if in == nil {
		return nil
	}
	out := new(PipelineIdentifier)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PipelineList) DeepCopyInto(out *PipelineList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Pipeline, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PipelineList.
func (in *PipelineList) DeepCopy() *PipelineList {
	if in == nil {
		return nil
	}
	out := new(PipelineList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PipelineList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PipelineSpec) DeepCopyInto(out *PipelineSpec) {
	*out = *in
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]apis.NamedValue, len(*in))
		copy(*out, *in)
	}
	if in.BeamArgs != nil {
		in, out := &in.BeamArgs, &out.BeamArgs
		*out = make([]apis.NamedValue, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PipelineSpec.
func (in *PipelineSpec) DeepCopy() *PipelineSpec {
	if in == nil {
		return nil
	}
	out := new(PipelineSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProviderAndId) DeepCopyInto(out *ProviderAndId) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProviderAndId.
func (in *ProviderAndId) DeepCopy() *ProviderAndId {
	if in == nil {
		return nil
	}
	out := new(ProviderAndId)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Run) DeepCopyInto(out *Run) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Run.
func (in *Run) DeepCopy() *Run {
	if in == nil {
		return nil
	}
	out := new(Run)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Run) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunConfiguration) DeepCopyInto(out *RunConfiguration) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunConfiguration.
func (in *RunConfiguration) DeepCopy() *RunConfiguration {
	if in == nil {
		return nil
	}
	out := new(RunConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RunConfiguration) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunConfigurationConversionRemainder) DeepCopyInto(out *RunConfigurationConversionRemainder) {
	*out = *in
	in.RunConversionRemainder.DeepCopyInto(&out.RunConversionRemainder)
	in.Triggers.DeepCopyInto(&out.Triggers)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunConfigurationConversionRemainder.
func (in *RunConfigurationConversionRemainder) DeepCopy() *RunConfigurationConversionRemainder {
	if in == nil {
		return nil
	}
	out := new(RunConfigurationConversionRemainder)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunConfigurationList) DeepCopyInto(out *RunConfigurationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RunConfiguration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunConfigurationList.
func (in *RunConfigurationList) DeepCopy() *RunConfigurationList {
	if in == nil {
		return nil
	}
	out := new(RunConfigurationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RunConfigurationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunConfigurationRef) DeepCopyInto(out *RunConfigurationRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunConfigurationRef.
func (in *RunConfigurationRef) DeepCopy() *RunConfigurationRef {
	if in == nil {
		return nil
	}
	out := new(RunConfigurationRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunConfigurationSpec) DeepCopyInto(out *RunConfigurationSpec) {
	*out = *in
	in.Run.DeepCopyInto(&out.Run)
	in.Triggers.DeepCopyInto(&out.Triggers)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunConfigurationSpec.
func (in *RunConfigurationSpec) DeepCopy() *RunConfigurationSpec {
	if in == nil {
		return nil
	}
	out := new(RunConfigurationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunConfigurationStatus) DeepCopyInto(out *RunConfigurationStatus) {
	*out = *in
	in.LatestRuns.DeepCopyInto(&out.LatestRuns)
	in.Dependencies.DeepCopyInto(&out.Dependencies)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunConfigurationStatus.
func (in *RunConfigurationStatus) DeepCopy() *RunConfigurationStatus {
	if in == nil {
		return nil
	}
	out := new(RunConfigurationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunConversionRemainder) DeepCopyInto(out *RunConversionRemainder) {
	*out = *in
	if in.Artifacts != nil {
		in, out := &in.Artifacts, &out.Artifacts
		*out = make([]OutputArtifact, len(*in))
		copy(*out, *in)
	}
	if in.ValueFromParameters != nil {
		in, out := &in.ValueFromParameters, &out.ValueFromParameters
		*out = make([]RuntimeParameter, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunConversionRemainder.
func (in *RunConversionRemainder) DeepCopy() *RunConversionRemainder {
	if in == nil {
		return nil
	}
	out := new(RunConversionRemainder)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunList) DeepCopyInto(out *RunList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Run, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunList.
func (in *RunList) DeepCopy() *RunList {
	if in == nil {
		return nil
	}
	out := new(RunList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RunList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunReference) DeepCopyInto(out *RunReference) {
	*out = *in
	if in.Artifacts != nil {
		in, out := &in.Artifacts, &out.Artifacts
		*out = make([]common.Artifact, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunReference.
func (in *RunReference) DeepCopy() *RunReference {
	if in == nil {
		return nil
	}
	out := new(RunReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunSchedule) DeepCopyInto(out *RunSchedule) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunSchedule.
func (in *RunSchedule) DeepCopy() *RunSchedule {
	if in == nil {
		return nil
	}
	out := new(RunSchedule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RunSchedule) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunScheduleList) DeepCopyInto(out *RunScheduleList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RunSchedule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunScheduleList.
func (in *RunScheduleList) DeepCopy() *RunScheduleList {
	if in == nil {
		return nil
	}
	out := new(RunScheduleList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RunScheduleList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunScheduleSpec) DeepCopyInto(out *RunScheduleSpec) {
	*out = *in
	in.RunSpec.DeepCopyInto(&out.RunSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunScheduleSpec.
func (in *RunScheduleSpec) DeepCopy() *RunScheduleSpec {
	if in == nil {
		return nil
	}
	out := new(RunScheduleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunSpec) DeepCopyInto(out *RunSpec) {
	*out = *in
	out.Pipeline = in.Pipeline
	if in.RuntimeParameters != nil {
		in, out := &in.RuntimeParameters, &out.RuntimeParameters
		*out = make([]RuntimeParameter, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Artifacts != nil {
		in, out := &in.Artifacts, &out.Artifacts
		*out = make([]OutputArtifact, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunSpec.
func (in *RunSpec) DeepCopy() *RunSpec {
	if in == nil {
		return nil
	}
	out := new(RunSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunStatus) DeepCopyInto(out *RunStatus) {
	*out = *in
	out.Status = in.Status
	in.Dependencies.DeepCopyInto(&out.Dependencies)
	if in.MarkedCompletedAt != nil {
		in, out := &in.MarkedCompletedAt, &out.MarkedCompletedAt
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunStatus.
func (in *RunStatus) DeepCopy() *RunStatus {
	if in == nil {
		return nil
	}
	out := new(RunStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RuntimeParameter) DeepCopyInto(out *RuntimeParameter) {
	*out = *in
	if in.ValueFrom != nil {
		in, out := &in.ValueFrom, &out.ValueFrom
		*out = new(ValueFrom)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RuntimeParameter.
func (in *RuntimeParameter) DeepCopy() *RuntimeParameter {
	if in == nil {
		return nil
	}
	out := new(RuntimeParameter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Status) DeepCopyInto(out *Status) {
	*out = *in
	out.ProviderId = in.ProviderId
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Status.
func (in *Status) DeepCopy() *Status {
	if in == nil {
		return nil
	}
	out := new(Status)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Triggers) DeepCopyInto(out *Triggers) {
	*out = *in
	if in.Schedules != nil {
		in, out := &in.Schedules, &out.Schedules
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.OnChange != nil {
		in, out := &in.OnChange, &out.OnChange
		*out = make([]OnChangeType, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Triggers.
func (in *Triggers) DeepCopy() *Triggers {
	if in == nil {
		return nil
	}
	out := new(Triggers)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ValueFrom) DeepCopyInto(out *ValueFrom) {
	*out = *in
	out.RunConfigurationRef = in.RunConfigurationRef
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ValueFrom.
func (in *ValueFrom) DeepCopy() *ValueFrom {
	if in == nil {
		return nil
	}
	out := new(ValueFrom)
	in.DeepCopyInto(out)
	return out
}
