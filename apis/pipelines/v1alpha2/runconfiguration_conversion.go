package v1alpha2

import (
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

type RunConfigurationConversionRemainder struct {
	RuntimeParameters []apis.NamedValue `json:"runtimeParameters,omitempty"`
}

func (rcr RunConfigurationConversionRemainder) empty() bool {
	return rcr.RuntimeParameters == nil
}

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunConfiguration)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters = mapToNamedValues(src.Spec.RuntimeParameters)
	dst.Spec.Pipeline = hub.PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	dst.Spec.Schedule = src.Spec.Schedule
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Status = hub.RunConfigurationStatus{
		Status: hub.Status{
			ProviderId:           src.Status.KfpId,
			SynchronizationState: src.Status.SynchronizationState,
			Version:              src.Status.Version,
			ObservedGeneration:   src.Status.ObservedGeneration,
		},
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
	}

	remainder := RunConfigurationConversionRemainder{}
	if err := retrieveAndUnsetConversionAnnotations(dst, &remainder); err != nil {
		return err
	}
	dst.Spec.RuntimeParameters = append(dst.Spec.RuntimeParameters, remainder.RuntimeParameters...)

	return nil
}

func (dst *RunConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunConfiguration)

	remainder := RunConfigurationConversionRemainder{}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters, remainder.RuntimeParameters = namedValuesToMap(src.Spec.RuntimeParameters)
	dst.Spec.Pipeline = PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	dst.Spec.Schedule = src.Spec.Schedule
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Status = RunConfigurationStatus{
		Status: Status{
			KfpId:                src.Status.ProviderId,
			SynchronizationState: src.Status.SynchronizationState,
			Version:              src.Status.Version,
			ObservedGeneration:   src.Status.ObservedGeneration,
		},
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
	}

	if err := setConversionAnnotations(dst, remainder); err != nil {
		return err
	}

	return nil
}
