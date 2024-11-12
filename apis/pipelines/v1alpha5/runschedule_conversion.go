package v1alpha5

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunSchedule) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunSchedule)
	v1alpha6Remainder := hub.RunScheduleConversionRemainder{}
	if err := pipelines.GetAndUnsetConversionAnnotations(src, &v1alpha6Remainder); err != nil {
		return err
	}
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Provider = getProviderAnnotation(src)
	removeProviderAnnotation(dst)
	dst.Spec.Pipeline = hub.PipelineIdentifier{
		Name:    src.Spec.Pipeline.Name,
		Version: src.Spec.Pipeline.Version,
	}
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Spec.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Artifacts = convertArtifactsTo(src.Spec.Artifacts)
	dst.Spec.Schedule = convertScheduleTo(
		src.Spec.Schedule,
		v1alpha6Remainder.Schedule,
	)
	dst.Status = hub.Status{
		ProviderId: hub.ProviderAndId{
			Provider: src.Status.ProviderId.Provider,
			Id:       src.Status.ProviderId.Id,
		},
		SynchronizationState: src.Status.SynchronizationState,
		Version:              src.Status.Version,
		ObservedGeneration:   src.Status.ObservedGeneration,
		Conditions:           hub.Conditions(src.Status.Conditions),
	}
	return nil
}

func (dst *RunSchedule) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunSchedule)
	v1alpha6Remainder := hub.RunScheduleConversionRemainder{}
	dst.ObjectMeta = src.ObjectMeta
	setProviderAnnotation(src.Spec.Provider, &dst.ObjectMeta)
	dst.Spec.Pipeline = PipelineIdentifier{
		Name:    src.Spec.Pipeline.Name,
		Version: src.Spec.Pipeline.Version,
	}
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Spec.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Artifacts = convertArtifactsFrom(src.Spec.Artifacts)
	schedule, err := convertCronExpressionFrom(src.Spec.Schedule, &v1alpha6Remainder)
	if err != nil {
		return err
	}
	dst.Spec.Schedule = schedule
	dst.Status = Status{
		ProviderId: ProviderAndId{
			Provider: src.Status.ProviderId.Provider,
			Id:       src.Status.ProviderId.Id,
		},
		SynchronizationState: src.Status.SynchronizationState,
		Version:              src.Status.Version,
		ObservedGeneration:   src.Status.ObservedGeneration,
		Conditions:           Conditions(src.Status.Conditions),
	}
	return pipelines.SetConversionAnnotations(dst, &v1alpha6Remainder)
}

// +kubebuilder:object:generate=false
type ConversionError struct {
	Message string
}

func (e *ConversionError) Error() string {
	return fmt.Sprintf("Error during conversion: %s", e.Message)
}

func convertCronExpressionFrom(
	schedule hub.Schedule,
	remainder *hub.RunScheduleConversionRemainder,
) (string, error) {
	if remainder == nil {
		return "", &ConversionError{"expected a v1alpha6 remainder but got nil"}
	}
	remainder.Schedule = schedule
	return schedule.CronExpression, nil
}

func convertScheduleTo(
	schedule string,
	remainder hub.Schedule,
) (hubSchedule hub.Schedule) {
	return hub.Schedule{
		CronExpression: schedule,
		StartTime:      remainder.StartTime,
		EndTime:        remainder.EndTime,
	}
}
